#!/usr/bin/env bash
# vim: et sw=4 ts=4 autoindent
#
# Copyright 2020 Hewlett Packard Enterprise Development LP
#
# Create a bootable pre-install-toolkit LiveCD USB drive.
#
# Script writes a hyrid ISO file to USB drive, creating two partitions
# during the write. After the ISO is written, Two additional partitions
# are then created: one for a copy-on-write (COW) data; second for
# install data.
#
# The filesystem on the COW partition is used by the system booting the
# ISO image as the overlayfs filesystem to store persistent changes to
# the root filesystem.
#
# The filesystem on the install data partition is used to store
# data needed to configure nodes as they are brought up and installed.
#
#   Partition   Description
#   ---------   -----------
#       1       CDROM partition with iso9660 filesystem
#       2       EFI partition
#       3       COW partition (ext4 filesystem, 'cow' label)
#       4       Install Data partition (ext4 filesystem, 'PITDATA' label)
#----------------------------------------------------------------------------
name=$(basename $0)

# Size in MB to use for cow partition
cow_size=0

# Size of specified block device
dev_size=0

# URL or path for ISO. Defaults to latest ISO build on master branch

iso_uri="http://$car/artifactory/csm/MTL/sle15_sp2_ncn/x86_64/dev/master/metal-team/cray-pre-install-toolkit-latest.iso"

# Initial empty values for usb device and iso file
usb=""
iso_file=""



usage () {
    cat << EOF
Usage $name USB-DEVICE ISO-FILE [COW-SIZE]

where:
    USB-DEVICE  Raw device file of USB flash drive where the CRAY-pre-install-toolkit (LiveCD)
                will be written.

    ISO-FILE    Pathname or URL of LiveCD ISO file to write to the usb
                flash drive.

    COW-SIZE    Size of the Copy on Write partition to create,(value in
                MB). If not specified the default value of 500 is used.
EOF
}

error () {
    mesg ERROR $@
}

warning () {
    mesg WARNING $@
}

info () {
    mesg INFO $@
}

mesg () {
    LEVEL=$1
    shift 1
    echo "$LEVEL: $@"
}

create_partition () {
    local part=$1
    local label=$2
    local dev=$3
    local start=$4
    local size=$5

    local dev_part=${dev}${part}
    local end_num=0


    if [[ $size > 0 ]]; then
        # Use specified size
        ((end_num=start+size))
    else
        # Use all remaining space on device
        ((end_num=dev_size - 1 ))
    fi

    if [[ $end_num -ge $dev_size ]]; then
        error "Not enough space to create ${dev_part} of size {$size}MB"
        exit 1
    fi

    info "Creating partition ${dev_part} for ${label} data: ${start}MB to ${end_num}MB"
    parted --wipesignatures -s $dev unit MB mkpart primary ext4 ${start}MB ${end_num}MB
    [[ $? -ne 0 ]] && error "Failed to create partition ${dev_part}" && exit 1

    # Wait for the partitioning and device file creation to complete.
    # Spin until file command is successful or too many atttempts.
    retcode=1
    tries=5
    while [[ $retcode -ne 0 ]]; do
        file ${dev_part} > /dev/null
        retcode=$?
        ((tries--))
        [[ $tries -eq 0 ]] && error "Failed to access partition ${dev_part}." && exit 1
        info "Waiting on ${dev_part} creation to complete"
        sleep 1
    done

    info "Making ext4 filesystem on partition ${dev_part}"
    mke2fs -L ${label} -t ext4 ${dev_part}
    [[ $? -ne 0 ]] && error "Failed to make filesystem on ${dev_part}" && exit 1
}

unmount_partitions () {
    local dev=$1

    # Check for device partitions that are mounted
    readarray -t mount_list < <(mount | egrep "${dev}[0-9]+" | awk '{print $1,$3}')
    if [[ ${#mount_list[@]} != 0 ]]; then
        echo "The following partition on ${dev} are mounted:"
        for i in ${!mount_list[@]}; do
            echo "    ${mount_list[$i]}"
        done
        error "Please unmount before attempting to run this format script again."
        exit 5
    fi
}

# Process cmdline arguments
[[ $# < 2 ]] && usage && exit 1
[[ $# > 3 ]] && usage && exit 1
usb=$1
shift 1
iso_file=$1
shift 1
if [[ $# -eq 0 ]]; then
    cow_size=500
else
    cow_size=$1
    shift 1
fi

# Validate the cow size is an integer > 1
if [[ $cow_size =~ [^0-9] ]]; then
    error "COW partition size was not specified as a number."
    echo ""
    usage
    exit 1
else
    if [[ $cow_size -lt 1 ]]; then
        error "COW partition must be at least 1MB in size."
        echo ""
        usage
        exit 1
    fi
fi


info "USB-DEVICE: $usb"
info "ISO-FILE:   $iso_file"
info "COW-SIZE:   ${cow_size}MB"

# Check to ensure the device exists
disk=${usb##*/}
if [[ $(lsblk | egrep "^${disk} " | wc -l) == 0 ]]; then
    error "Device ${usb} not found via lsblk."
    exit 1
fi

# check to ensure the ISO file exists
if [[ ! -r "$iso_file" ]]; then
    error "File ${iso_file} does not exist or is not readable."
    exit 1
fi

unmount_partitions $usb

# Check the downloaded ISO to ensure it is valid. Better
# to know now vs. when it fails the checkmedia during boot.
# Emit a warning if the command is not there
if ! eval command -v checkmedia; then
  error "Unable to validate ISO.  Please install checkmedia and try again."
  exit 4
else
  info "Validating ISO via checkmedia"
  checkmedia $iso_file
  [[ $? -ne 0 ]] && error "Failed checkmedia verification of $iso_file" && exit 1
fi

# Write new partition table
info "Writing new GUID partition table to ${usb}"
parted -s $usb mktable gpt

# Write the ISO to the USB raw device, creating an exact duplicate
# of the ISO image layout.
info "Writing ISO to $usb"
dd bs=1M if=$iso_file of=$usb conv=fdatasync
[[ $? -ne 0 ]] && error "Failed to write $iso_file to $usb" && exit 1

info "Scanning $usb for where to begin creating partition"
readarray -t parted_line < <(parted -s -m $usb unit MB print)
start_num=0
end_num=0
part_num=0
for i in ${!parted_line[@]}; do

    # Parse the line into fields
    IFS=":" read -r -a fields <<< ${parted_line[$i]}

    # Line beginning with USB device name gives total size of drive
    if [[ ${fields[0]} == ${usb} ]]; then
        # Get disk dev size
        dev_size=${fields[1]%%MB*}


    # Error if three partitions found, no space for cow
    # and install data partitions
    elif [[ ${fields[0]} == 3 ]]; then
        error "Found 3 partitions, no partition left for install data"
        exit 1

    # Find end of each existing partition, install data partition will
    # begin after the end of the last partition.
    elif [[ ${fields[0]} == [12] ]]; then

        # Get end of partition
        start_num=${fields[2]%%MB*}

        # Start partition after end of last found partition
        ((start_num++))

        # Track what number next partition will be
        part_num=${fields[0]}
        ((part_num++))
    fi
done

# Create cow partition for liveCD
create_partition $part_num "cow" $usb $start_num $cow_size

# Create the install data partition for configuration data using
# remaining space
((part_num++))
((start_num=start_num+cow_size+1))
create_partition $part_num "PITDATA" $usb $start_num 0

info "Partition table for $usb"
parted -s $usb unit MB print
