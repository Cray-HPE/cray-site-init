@Library("dst-shared@master") _
rpmBuild(
    specfile : "cray-site-init.spec",
    product : "csm",
    target_node : "ncn",
    fanout_params : ["sle15sp2"],
    channel : "csi-ci-alerts",
    slack_notify : ['', 'SUCCESS', 'FAILURE', 'FIXED']
)
