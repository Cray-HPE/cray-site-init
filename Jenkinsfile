@Library("dst-shared@master") _
rpmBuild(
    githubPushRepo: "Cray-HPE/cray-site-init",
    githubPushBranches: "release/.*|main",
    master_branch: "main",
    specfile : "cray-site-init.spec",
    product : "csm",
    target_node : "ncn",
    fanout_params : ["sle15sp2"],
    channel : "csi-ci-alerts",
    slack_notify : ['', 'SUCCESS', 'FAILURE', 'FIXED']
)
