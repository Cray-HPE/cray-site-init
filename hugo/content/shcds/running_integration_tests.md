---
layout: default
title: Running integration tests with canu and csi
---

In CSM 1.2+ `csi` will have the ability to ingest a JSON file and automatically generate the seed files needed for the `csi config init` command.

## Seed Files

`csi config init` relies on several "seed" files in oder to produce a Shasta configuration payload:

- `hmn_connections.json`: maps which switch ports on the Leaf switches are cabled to what the River node BMCs, PDUs, or other hardware
- `ncn_metadata.csv`: maps the MACs of the Management NCNs to their xnames and is used to initialize the management cluster
- `switch_metdata.csv`: maps the switch xname, brand, type, and model for the management switches in the system
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
