@Library("dst-shared@release/shasta-1.4") _
rpmBuild(
    specfile : "cray-site-init.spec",
    product : "csm",
    target_node : "ncn",
    fanout_params : ["sle15sp2"],
    channel : "csi-ci-alerts",
    slack_notify : ['', 'SUCCESS', 'FAILURE', 'FIXED']
)
