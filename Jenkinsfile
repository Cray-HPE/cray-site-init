@Library("dst-shared@master") _
rpmBuild(
    specfile : "shasta-instance-control.spec",
    product : "shasta-premium",
    target_node : "ncn",
    fanout_params : ["sle15sp2"],
    channel : "metal-ci-alerts",
    slack_notify : ['', 'SUCCESS', 'FAILURE', 'FIXED']
)
