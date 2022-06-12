---
layout: default
title: Running integration tests with canu and csi
---

The SHCD is the beginning source-of-truth for how a system is laid out and connected.  The information found there is used through several different tools throughout the install of CSM, so if changes are made to an SHCD, it can be beneficial in both development and production environments to see how those changes might propagate.

To that end, there are some integration tests built in to `csi` that can test the flow mentioned above, which allow you to generate seed files for multiple systems at once.

# High-level overview

1. Add `shcd.xlsx` files into `testdata/shcds`
2. Add testdata
2. Run `make shcds` 
3. Watch the files get created in `testdata/shcds`

## Add `shcd.xlsx` files into `testdata/shcds`

Put a copy of each `shcd.xlsx` you want to generate files for into `testdata/shcds`.

```
~/cray-site-init $ ls -l testdata/shcds/
total 68120
-rw-r--r--@ 1 seymour  12790  4957149 Oct 27 08:41 HPE System 598 Shasta River Odin_RevA22.xlsx
-rw-r--r--@ 1 seymour  12790  5177069 Nov  2 13:13 Hela.xlsx
-rw-r--r--@ 1 seymour  12790  3877457 Nov  1 12:35 Shasta River Drax CCD.xlsx
-rw-r--r--@ 1 seymour  12790  1684315 Nov  1 12:45 Shasta River Fanta CCD.xlsx
-rw-r--r--@ 1 seymour  12790  1672018 Nov  2 13:23 Shasta River Redbull SHCD.xlsx
-rw-r--r--@ 1 seymour  12790  3491609 Nov  2 13:39 Shasta River Rocket SHCD.xlsx
-rw-r--r--@ 1 seymour  12790  2669723 Nov  2 13:15 Shasta River System_Ego_CCD.xlsx
-rw-r--r--@ 1 seymour  12790  1677461 Nov  1 13:46 Shasta River Thanos SHCD.xlsx
-rw-r--r--@ 1 seymour  12790  3896541 Nov  1 15:05 System3 Baldar Shasta  River.xlsx
-rw-r--r--@ 1 seymour  12790  2744277 Nov  1 13:57 System4 Sif Shasta River.xlsx
-rw-r--r--@ 1 seymour  12790  3007754 Nov  1 15:36 System6 Loki Shasta  River  SHCD.xlsx
```

## Add testdata

`canu` requires several arguments in order to generate an `shcd.json`.  The necessary arguments can be added to the `tests` struct in `cmd/shcd_integration_test.go` similar to the example below.

```
var canus = []struct {
	systemName string
	version    string
	tabs       string
	corners    string
}{
	{
		systemName: "mysupercomputer",
		version:    "V1",
		tabs:       "25G_10G,NMN,HMN",
		corners:    "I14,S65,J16,T23,J20,U46",
	},
```

## Run `make shcds`

```bash
# make sure canu is installed
make shcds
```

## Watch the files get created in `testdata/shcds`

On a successful test, you'll see output similar to:

```
=== RUN   TestConfigShcd_GenerateSeeds
2021/11/02 12:47:54 Running canu to create shcd.json for baldar...
2021/11/02 12:47:56 Using schema file: ../internal/files/shcd-schema.json
2021/11/02 12:47:56 Created switch_metadata.csv from SHCD data
2021/11/02 12:47:56 Created hmn_connections.json from SHCD data
2021/11/02 12:47:56 Running canu to create shcd.json for drax...
2021/11/02 12:47:58 Using schema file: ../internal/files/shcd-schema.json
2021/11/02 12:47:58 Created switch_metadata.csv from SHCD data
2021/11/02 12:47:56 Created hmn_connections.json from SHCD data
2021/11/02 12:47:58 Running canu to create shcd.json for ego...
...
...
```

`testdata/shcds/` will also have folders matching each of the system names and contain the following files:

- `application_node_config.yaml`
- `shcd.json`
- `switch_metadata.csv`
- `hmn_connections.json`
