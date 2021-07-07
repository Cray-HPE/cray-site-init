---
date: 2021-07-07T16:41:32-05:00
title: "csi pit populate pitdata"
layout: default
---
## csi pit populate pitdata

Populates the PITDATA partition with necessary config files

### Synopsis

Populates the PITDATA partition with necessary config files.

	SRC can be a path to a folder with squashfs image (-k and -c flags).
	SRC can be a path to a folder of csi-generated files (-b flag)
	SRC can be a path to any folder where you only want files copied over (-p flag)
	DEST should be a path to where KIS components will be copied to

```
csi pit populate pitdata SRC DEST [flags]
```

### Options

```
  -b, --basecamp     Copy any discovered basecamp config files to the 'configs' directory on the PITDATA partition
  -C, --ceph         Copy any discovered ceph squashfs images from SRC to DEST
  -h, --help         help for pitdata
  -i, --initrd       Copy any discovered initrds from SRC to DEST
  -k, --kernel       Copy any discovered kernels from SRC to DEST
  -K, --kubernetes   Copy any discovered k8s squashfs images from SRC to DEST
  -p, --prep         Copy only files from a directory from SRC to DEST
```

### SEE ALSO

* [csi pit populate](/commands/csi_pit_populate/)	 - Populates the LiveCD with configs

