---
layout: default
title: Using a canu-generated shcd.json with csi
---

In CSM 1.2+ `csi` will have the ability to ingest a JSON file and automatically generate the seed files needed for the `csi config init` command.

## Seed Files

`csi config init` relies on several "seed" files in oder to produce a Shasta configuration payload:

- `hmn_connections.json`: maps which switch ports on the Leaf switches are cabled to what the River node BMCs, PDUs, or other hardware
- `ncn_metadata.csv`: maps the MACs of the Management NCNs to their xnames and is used to initialize the management cluster
- `switch_metadata.csv`: maps the switch xname, brand, type, and model for the management switches in the system
- (optional) `application_node_config.yaml`: offers additional control of the application node identification in SLS
- (optional) `cabinets.yaml`: a mapping file is necessary for systems with non-sequential cabinet ID numbers

The three non-optional files represent the _minimum_ set of inputs `csi` needs to generate a new configuration payload.  

### Auto-generating the Seed Files

In older versions of CSM, these seed files needed to be hand-crafted, which was painful and error-prone.  This process has improved some in CSM v1.2 with the help of [`canu`](https://github.com/Cray-HPE/canu).

You can automatically generate the seed files in `csi` _after_ you create a `shcd.json` using `canu`.  This JSON file is a machine-readable version the SHCD.

Here is an example workflow:

```bash
# Generate a machine-readable JSON file from an Excel spreadsheet
canu validate shcd --shcd MySystem_SHCD.xlsx --tabs HMN --corners j20,u53 -a v1 --out MySystem.json

# Use the machine-readable JSON with csi to auto-generate the seed files
# In this example, the hmn_connections.json file and the switch_metadata.csv are being generated
# This allows you to generate the entire set of seed files or just certain ones
csi config shcd --hmn-connections --switch-metadata MySystem.json

# Use the newly-generated files with the traditional workflow
csi config init
```

# Running integration tests with canu and csi

The SHCD is the beginning source of truth for how a system is laid out and connected.  The information found there is used through several different tools throughout the install of CSM, so if changes are made to an SHCD, it can be beneficial in both development and production environments to see how those changes might propagate.

To that end, there are some integration tests built in to `csi` that can test the flow mentioned above. 


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

Once that test data is in place,

```bash
mkdir testdata/shcds
# ensure SHCD spreadsheets are in testdata/shcds and the system name is somewhere in the filename
cp mysupercomputer.xlsx testdata/shcds
# make sure canu is installed
make shcds
```

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
