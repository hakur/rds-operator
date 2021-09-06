#!/bin/bash

function BackupMaster() {
    local master=$(FindMasterServer)
    local outputFile=$1

    if [ "$master" == "" ];then
        echolog "master node not found"
        return 1
    fi

    local cmd=""
    local fileExt="zlib"
    if [ "$MYSQL_BACKUP_USE_ZLIB_COMPRESS" == "true" ];then
        cmd="--compress-output=ZLIB "
        cmd=$cmd"--single-transaction"
    else
        cmd="--single-transaction"
        fileExt="sql"
    fi

    mysqlpump -u$MYSQL_USER  -P$MYSQL_PORT -h$master \
    --all-databases \
    --add-drop-user \
    --add-drop-table \
    --add-drop-database \
    --default-parallelism=$(nproc) \
    --skip-watch-progress \
    $cmd > ${outputFile}_from_${master}.${fileExt}
}

function UploadToS3() {
    local scanDirPath=$1
    mc alias set auth $S3_ENDPOINT $S3_ACCESS_KEY $S3_SECURITY_KEY
    ls $scanDirPath | while read file; do
        mc cp $scanDirPath/$file auth/$S3_BUCKET/$S3_PATH/$file
    done
}

function DownloadFileFromS3() {
    local s3Filepath=""
    local outputFilepath=""
    mc alias set auth $S3_ENDPOINT $S3_ACCESS_KEY $S3_SECURITY_KEY
    mc cat auth/$S3_BUCKET/$s3Filepath > $outputFilepath
}