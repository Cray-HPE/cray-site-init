@Library('dst-shared@master') _
dockerBuildPipeline {
     app = "sic"
     name = "shasta-install-control"
     description = "Container for controlling the state of a Shasta install on metal or virtual."
     dockerfile = "Dockerfile"
     repository = "metal"
     imagePrefix = "cloud"
     product = "shasta-standard,shasta-premium"
     slackNotification = ["", "", false, true, true, true]
}
