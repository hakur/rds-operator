#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail
rm -rf ../clientset
rm -rf ../listers
rm -rf ../informers
../../code-generator/generate-groups.sh "all" "github.com/hakur/rds-operator" "github.com/hakur/rds-operator" "apis:v1alpha1" --output-base "" --go-header-file "./boilerplate.go.txt"

mv github.com/hakur/rds-operator/clientset ..
mv github.com/hakur/rds-operator/informers ..
mv github.com/hakur/rds-operator/listers ..

rm -rf github.com
