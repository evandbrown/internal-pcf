wget https://bosh.io/d/github.com/concourse/concourse?v=3.14.1
wget https://bosh.io/d/github.com/cloudfoundry/garden-runc-release?v=1.14.0
wget https://bosh.io/d/github.com/cloudfoundry/postgres-release?v=28

mv 'concourse?v=3.14.1' concourse.release
mv 'garden-runc-release?v=1.14.0' garden-runc-release.release
mv 'postgres-release?v=28' postgres-release.release

bosh deploy -d concourse concourse.yml \
  -l ../versions.yml \
  --vars-store cluster-creds.yml \
  -o operations/no-auth.yml \
  -o operations/no-internet-access.yml \
  -o operations/privileged-http.yml \
  -o operations/privileged-https.yml \
  -o operations/tls.yml \
  -o operations/tls-vars.yml \
  -o operations/web-network-extension.yml \
  --var network_name=default \
  --var web_vm_type=default \
  --var db_vm_type=default \
  --var db_persistent_disk_type=10GB \
  --var worker_vm_type=default \
  --var deployment_name=concourse \
  --var web_network_name=private \
  --var web_network_vm_extension=ilb \
  --var atc_basic_auth.username=admin \
  --var atc_basic_auth.password=admin \
  --var concourse_release=concourse.release \
  --var garden_runc_release=garden-runc-release.release \
  --var postgres_release=postgres-release.release \
  --var external_url=https://10.100.0.100
