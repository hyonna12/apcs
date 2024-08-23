pipeline{

    agent any

        environment {

            // Preparation Stage
            WORK_DIR = "/home/apcs"
            GITLAB_ADDR="https://${GITLAB_ID}:${GITLAB_PW}@gitlab.mipllab.com/lw/ap/apcs.git"
            GITLAB_BRANCH="dev_refactored"
            
            // Build Stage
            HARBOR_ADDR="harbor.mipllab.com -u ${HARBOR_ID} -p ${HARBOR_PW}"
            IMG_1="apcs"
            TAG_1="harbor.mipllab.com/lw/apcs:latest"


            // Deployment Stage
            DEPLOY_DIR = "/home/tomcat/docker/apcs"
            // APP_PATH_1="/home/target/k8s/app.yaml"
            DOCKER_COMPOSE_FILE="${WORK_DIR}/docker-compose.yaml"

        }

    stages{



        stage ('Preparation'){

            steps {

                sh ''' #!/bin/bash

                    mkdir -p ${WORK_DIR}

                    git -C ${WORK_DIR}/ init

                    git -C ${WORK_DIR}/ pull ${GITLAB_ADDR} ${GITLAB_BRANCH}

                '''

            }

        }




        stage ('Build'){

            steps {

                sh ''' #!/bin/bash

                    docker login ${HARBOR_ADDR}

                    cd ${WORK_DIR} && docker build -t ${TAG_1} .

                    docker push ${TAG_1}

                '''

            }

        }

        stage ('Deployment'){

            steps {

                sh ''' #!/bin/bash

                    ssh -p 9022 tomcat@lineworldap.iptime.org "mkdir -p ${DEPLOY_DIR}"

                    scp -P 9022 ${DOCKER_COMPOSE_FILE} tomcat@lineworldap.iptime.org:${DEPLOY_DIR}

                    ssh -p 9022 tomcat@lineworldap.iptime.org "cd ${DEPLOY_DIR} && docker compose down && docker rmi -f ${TAG_1}"

                    ssh -p 9022 tomcat@lineworldap.iptime.org "cd ${DEPLOY_DIR} && docker logout && docker login ${HARBOR_ADDR}  && docker compose up -d --build"


                '''
            }

        }

        stage ('Termination'){

            steps {

                sh ''' #!/bin/bash

                    rm -rf ${WORK_DIR}

                    docker rmi ${TAG_1}

                '''


            }


        }
        



    }
}

