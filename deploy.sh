#!/bin/bash

VERSION=0.0.$TRAVIS_BUILD_NUMBER

# vars for testing. comment out for release.
#AWS_SECRET=
#AWS_BUCKET=
#AWS_KEY=
#TRAVIS_BRANCH=
#TRAVIS_BUILD_NUMBER=
#TRAVIS_COMMIT=

# copy files to s3
echo copy files to current release dir
AWS_ACCESS_KEY_ID=$AWS_KEY AWS_SECRET_ACCESS_KEY=$AWS_SECRET aws s3 cp \
    --recursive bin/$TRAVIS_BRANCH/ s3://release.gocms.io/$TRAVIS_BRANCH/current/

# copy files to build number bucket
AWS_ACCESS_KEY_ID=$AWS_KEY AWS_SECRET_ACCESS_KEY=$AWS_SECRET aws s3 cp \
    --recursive s3://release.gocms.io/$TRAVIS_BRANCH/current/ s3://release.gocms.io/$TRAVIS_BRANCH/$VERSION/