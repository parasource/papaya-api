properties([disableConcurrentBuilds()])

pipeline {
    agent {
        label 'jenkins_agent'
	}
    triggers { pollSCM('* * * * *') }
    options {
        buildDiscarder(logRotator(numToKeepStr: '10', artifactNumToKeepStr: '10'))
        timestamps()
    }
    stages {
        stage("docker login") {
            steps {
                echo " ============== docker login =================="
                withCredentials([usernamePassword(credentialsId: 'dockerhub_sireax', usernameVariable: 'sireax', passwordVariable: '65onion56')]) {
                    sh """
                    docker login -u $USERNAME -p $PASSWORD
                    """
                }
            }
        }
        stage("create docker image") {
            steps {
                echo " ============== start building image =================="
                dir ('docker/toolbox') {
                	sh 'docker build -t sireax/papaya-api:latest . '
                }
            }
        }
        stage("docker push") {
            steps {
                echo " ============== start pushing image =================="
                sh '''
                docker push sireax/papaya-api:latest
                '''
            }
        }
    }
}