#!/bin/bash

# bash script to install the flux controller

helm install -n flux-system flux oci://ghcr.io/fluxcd-community/charts/flux2 --create-namespace