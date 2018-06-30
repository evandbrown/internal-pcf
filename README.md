This document describes how to use the `internal-pcf` repository to:
1. Provision a Concourse installation that prioritizes private networking. A NAT VM and jumpbox are the only two resources that have public addresses.
1. Deploy a Pivotal Cloud Foundry installation (via a Concourse pipeline) that prioritizes private networking. A trio of NAT VMs and the Operations Manager VM are the only resources that have public addresses.

# Pre-work
1. Install the [BOSH Bootloader](https://github.com/cloudfoundry/bosh-bootloader) and its [dependencies](https://github.com/cloudfoundry/bosh-bootloader#install-dependencies).
1. Install `direnv`
1. Clone this repository

# Deploying Concourse
1. Create a new directory with the structure required to apply customizations:

    ```
    mkdir env-stage-internal-pcf && cd !$
    export BBL_DIR=$(pwd)
    ```

1. Initialize the directory as a git repository:

    ```
    git init
    ```

1. From the directory where you cloned this repository, run the following command to copy the customizations in:

    ```
    cd bbl-overlays
    cp .envrc ${BBL_DIR}
    cp byo-network/*.sh ${BBL_DIR}
    cp byo-network/*.yml ${BBL_DIR}
    for dir in byo-network internal-tcp-lb nat; do
      pushd ${dir}
        rsync -avzh --ignore-errors terraform ${BBL_DIR}
        rsync -avzh --ignore-errors vars ${BBL_DIR}
        rsync -avzh --ignore-errors cloud-config ${BBL_DIR}
      popd
    done
    ```

1. Go to ${BBL_DIR/vars and edit `.envrc`

1. Go to ${BBL_DIR}/vars and edit `byo_network.tfvars` and `internal_tcp_lb.tfvars`

1. Use `bbl` to provision a BOSH director:

    ```
    cd ${BBL_DIR}
    bbl plan --name <CHANGE>
    bbl up
    ```

1. Load the creds and upload a stemcell:

    ```
    eval "$(bbl print-env)"
    bosh upload-stemcell --sha1 7333cc0d2042cd7b167a7d57c03f38562cd4e01c \
      https://bosh.io/d/stemcells/bosh-google-kvm-ubuntu-trusty-go_agent?v=3586.24
    ```

1. In the original cloned repo, navigate to `concourse-bosh-deployment/cluster`, edit <CHANGE> values in `deploy.sh`, then run it.

# Misc
1. TODO: compilation VMs have public addresses
1. Setup socks5 proxy to jumpbox (e.g., ssh -i id_rsa -D 9999 -q -N jumpbox@35.227.150.58 -f)
  * In Crostini and in ChromeOS WiFi settings
