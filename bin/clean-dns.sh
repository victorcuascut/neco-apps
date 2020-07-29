#! /bin/bash -e

gcloud --project=neco-test dns record-sets export --zone-file-format --zone=gcp0 zone-gcp0
gcloud --project=neco-test dns record-sets transaction start --zone=gcp0
