@Library('dst-shared@master') _

dockerBuildPipeline {
        githubPushRepo = "Cray-HPE/hms-securestorage"
        repository = "cray"
        imagePrefix = "hms"
        app = "securestorage"
        name = "hms-securestorage"
        description = "Cray HMS securestorage code."
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "internal"
}
