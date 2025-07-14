---
date: 2025-07-10T15:00:19-05:00
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
  -c, --config string            Path to a CSI config file (default is $PWD/system_config.yaml).
      --csm-api-url string       (for use against a completed CSM installation) The URL to a CSM API. (default "https://api-gw-service-nmn.local")
  -i, --input-dir string         A directory to read input files from (--config will take precedence, but only for system_config.yaml).
      --k8s-namespace string     (for use against a completed CSM installation) The namespace that the --k8s-secret-name belongs to. (default "default")
      --k8s-secret-name string   (for use against a completed CSM installation) The name of the Kubernetes secret to look for an OpenID credential in for CSM APIs (a.k.a. TOKEN=). (default "admin-client-auth")
```

### SEE ALSO

* [csi](/commands/csi/)	 - Cray Site Initializer (csi)

