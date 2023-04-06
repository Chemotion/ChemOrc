#!/bin/bash

[[ -d /backup ]] || {
    log "Backup skipped. [/backup] is not mounted."
    exit 1
}

source /embed/lib/dbenv

BACKUP_WHAT=${BACKUP_WHAT:-both}

stamp=$(date +'%y%m%d-%H%M%S')

if [[ $BACKUP_WHAT == "both" || $BACKUP_WHAT == "data" ]]; then
    tar cvzH posix -f "/backup/backup-${stamp}.data.tar.gz" --directory=/chemotion/data/ . --directory=/ .version || {
        log "Could not backup user data!"
        exit 2
    }
fi

if [[ $BACKUP_WHAT == "both" || "$BACKUP_WHAT" = "db" ]]; then
    pg_dump --no-owner --clean --if-exists | gzip -c > "/backup/backup-${stamp}.sql.gz" || {
        log "Could not backup database!"
        exit 3
    }
fi


if [[ $BACKUP_WHAT == "both" || $BACKUP_WHAT == "data" ]]; then
    if [[ ! -e "/backup/backup.data.tar.gz" || -L "/backup/backup.data.tar.gz" ]]; then
        log "Creating symlink to latest data backup: /backup/backup.data.tar.gz -> backup-${stamp}.data.tar.gz"
        rm -f /backup/backup.data.tar.gz
        ln -s "backup-${stamp}.data.tar.gz" "/backup/backup.data.tar.gz"
    fi
fi

if [[ $BACKUP_WHAT == "both" || $BACKUP_WHAT == "db" ]]; then
    if [[ ! -e "/backup/backup.sql.gz" || -L "/backup/backup.sql.gz" ]]; then 
        log "Creating symlink to latest db backup: /backup/backup.sql.gz -> backup-${stamp}.sql.gz"
        rm -f /backup/backup.sql.gz
        ln -s "backup-${stamp}.sql.gz" "/backup/backup.sql.gz"
    fi
fi

log "Backup finished successfully. Timestamp: ${stamp}"
