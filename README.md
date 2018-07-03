This document describes how to use the `internal-pcf` repository to:
1. Provision a Concourse installation that prioritizes private networking. A NAT VM is the only resource that has a public address.
1. Deploy a Pivotal Cloud Foundry installation (via a Concourse pipeline) that prioritizes private networking. A trio of NAT VMs is the only resources that has a public address.

# Important assumptions
1. This project is focused on deploying into an existing networking environment and assumes that a VPC network and 5 subnetworks already exist. No new network or subnetworks will be created. 
1. **Really important:** There is no external ingress configured for this installation. You will either need an existing VPN tunnel/gateway configured into your VPC, or perform all of the below instructions from an accessible jumpbox in one of the existing subnetworks. Note that the jumpbox created by `bbl` does not have a public address.


# Pre-work
1. If you will be using existing subnetworks that are smaller than `/16`, you will need a patched version of `bbl` until [this PR](https://github.com/cloudfoundry/bosh-bootloader/pull/335) is merged. Download here: [Linux](https://storage.googleapis.com/evandbrown17/bbl-linux)(bf23eb4f95e07392858613809a0796bb1caf97d5) or [Mac](https://storage.googleapis.com/evandbrown17/bbl-linux)(d76f510decf402ea873f014449131590220374ae).
1. Install `bbl`'s [dependencies](https://github.com/cloudfoundry/bosh-bootloader#install-dependencies).
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
    for dir in byo-network internal-tcp-lb nat jumpbox_private_ip; do
      pushd ${dir}
        rsync -avzh --ignore-errors terraform ${BBL_DIR}
        rsync -avzh --ignore-errors vars ${BBL_DIR}
        rsync -avzh --ignore-errors cloud-config ${BBL_DIR}
      popd
    done
    ```

1. Go to ${BBL_DIR/vars and edit `.envrc`

1. Go to ${BBL_DIR}/vars and edit `byo_network.tfvars`. You must specify a subnet that is just for Concourse - the upcoming PCF installation can not use this subnet.

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

# Deploy Concourse

1. In the original cloned repo, navigate to `concourse-bosh-deployment/cluster`, edit <CHANGE> values in `deploy.sh`, then run it.

1. Once Concourse has been deployed, locate its load balancer's private address and use `fly` to target the installation.

# Deploy pcf-pipelines

1. In the original cloned repo, navigate to `pcf-pipelines/install-pcf/gcp` and edit the values in `params.yml`.

1. Edit `pipeline.yml` as follows:

  * set `source.uri` in the `pcf-pipelines` resource to `https://github.com/evandbrown/pcf-pipelines.git`,

  * set `source.branch` in the `pcf-pipelines` resource to `internal-pipelines`

  * remove `source.private_key`

  * the `pcf-pipelines` resource in your `pipelines.yml should look something like:

      ```
      ...
      resources:
      - name: pcf-pipelines
        type: git
        source:
          uri: https://github.com/evandbrown/pcf-pipelines.git
          branch: internal-pipelines
      ...

      ```

1. Apply the pipeline with something like `fly -t your-team set-pipeline -p your-pipeline-name -c pipeline.yml -l params.yml`

1. Navigate to your pipeline's webpage, unpause it, then manually trigger the `bootstrap-terraform-state` task, then manually trigger every task sequentially starting with `upload-opsman-image`


