---
date: 2024-06-06T09:39:24-05:00
title: "csi upload-sls-file"
layout: default
---
## csi upload-sls-file

Upload the sls_input_file.json file into the SLS S3 Bucket

### Synopsis

Upload the given sls_input_file.json file into the SLS S3 Bucket.
	Example: csi upload-sls-file --sls-file /path/to/sls_input_file.json

	Upload the given sls_input_file.json file into the SLS S3 Bucket, and delete the key "uploaded" if present in the SLS S3 bucket.
	The "upload" flag is added by the SLS loader when it successfully loads SLS with data, and it is used as flag to determine if
	the SLS loader has been used previously to upload the SLS file. If it is present the loader will not upload the SLS file.

	Example: csi upload-sls-file --sls-file /path/to/sls_input_file.json --remove-upload-flag
	

```
csi upload-sls-file [flags]
```

### Options

```
  -h, --help                 help for upload-sls-file
      --kubeconfig string    Absolute path to the kubeconfig file (default "/Users/rusty/.kube/config")
      --remove-upload-flag   Remove the upload flag added by the SLS loader
      --s3-bucket string     Bucket to create and upload the SLS input file to (default "sls")
      --s3-secret string     Secret to use for connecting to S3 (default "sls-s3-credentials")
      --sls-file string      Path to the SLS Input File to Upload (default "sls_input_file.json")
```

### Options inherited from parent commands

```
  -c, --config string   CSI config file
```

### SEE ALSO

* [csi](/commands/csi/)	 - Cray Site Init. For new sites, re-installs, and upgrades.

