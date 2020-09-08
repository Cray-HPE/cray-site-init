@Library('dst-shared@master') _
dockerBuildPipeline {
     app = "sic"
     name = "shasta-instance-control"
     description = "Container for controlling the state of a Shasta instance, on bare-metal or virtual platforms."
     dockerfile = "Dockerfile"
     repository = "metal"
     imagePrefix = "cloud"
     product = "shasta-standard,shasta-premium"
     slackNotification = ["", "", false, true, true, true]
}
