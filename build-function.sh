export BASE_FUNCTIONS_DIR=/home/oracle/go/go_projects/src/oci/dump_oci
##cd $BASE_FUNCTIONS_DIR/fn-adw-event-notification-app-v2
cd /home/oracle/BuildDocker/Oracle12
fn create context Eugene --provider oracle
fn use context Eugene
fn update context oracle.compartment-id ocid1.compartment.oc1..aaaaaaaa375sfgxnc24b3rmxjju6ttxv264t6ukiyv42txxfxs3zj2difroa
fn update context api-url https://functions.eu-frankfurt-1.oraclecloud.com
fn update context registry fra.ocir.io/oraseemeatechse/eugenesimos
cat $BASE_FUNCTIONS_DIR/token | docker login  fra.ocir.io --username 'oraseemeatechse/oracleidentitycloudservice/eugene.simos@oracle.com' --password-stdin
##cd $BASE_FUNCTIONS_DIR/fn-adw-event-notification-app-v2
fn delete function fncomputeapp notifysql
##docker system prune -f
docker rmi $(docker images -a -q) -f
fn -v deploy --app fncomputeapp  --no-cache