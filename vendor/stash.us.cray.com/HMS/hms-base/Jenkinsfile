@Library('dst-shared@master') _

dockerBuildPipeline {
        githubPushRepo = "Cray-HPE/hms-base"
        repository = "cray"
        imagePrefix = "hms"
        app = "base"
        name = "hms-base"
        description = "Cray HMS base code."
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "internal"
}
