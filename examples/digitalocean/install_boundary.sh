DB_PUBLIC_IP=$(terraform output -raw db_public_ip)
DB_PRIVATE_IP=$(terraform output -raw db_private_ip)

CONTROLLER_PUBLIC_IP=$(terraform output -raw controller_public_ip)
CONTROLLER_PRIVATE_IP=$(terraform output -raw controller_private_ip)

WORKER1_PUBLIC_IP=$(terraform output -raw worker1_public_ip)
WORKER1_PRIVATE_IP=$(terraform output -raw worker1_private_ip)

WORKER2_PUBLIC_IP=$(terraform output -raw worker2_public_ip)
WORKER2_PRIVATE_IP=$(terraform output -raw worker2_private_ip)

ROOT_KEY=JuZetUPcWEbsO4cTP7/E93O8Zrl6CQnYYb25CLFVAJM=
WORKER_AUTH_KEY=uDuFypRqVPq7DM0Ujgrl5PoRRgqqy4eFdW7ORNIHHhU=
RECOVERY_KEY=eixF4zdao9C6Soe+5jzBTtCnLUGMsP72NW5x30QP6So=

hashi-up boundary init-database \
    --ssh-target-addr $DB_PUBLIC_IP \
    --db-url "postgresql://boundary:boundary123@localhost:5432/boundary?sslmode=disable" \
    --root-key $ROOT_KEY

hashi-up boundary install \
    --ssh-target-addr $CONTROLLER_PUBLIC_IP \
    --controller-name do-boundary-ctl \
    --db-url "postgresql://boundary:boundary123@$DB_PRIVATE_IP:5432/boundary?sslmode=disable" \
    --root-key $ROOT_KEY \
    --worker-auth-key $WORKER_AUTH_KEY \
    --cluster-addr $CONTROLLER_PRIVATE_IP \
    --recovery-key $RECOVERY_KEY

hashi-up boundary install \
    --ssh-target-addr $WORKER1_PUBLIC_IP \
    --worker-name do-boundary-wkr-1 \
    --public-addr $WORKER1_PUBLIC_IP \
    --worker-auth-key $WORKER_AUTH_KEY \
    --controller $CONTROLLER_PRIVATE_IP

hashi-up boundary install \
    --ssh-target-addr $WORKER2_PUBLIC_IP \
    --worker-name do-boundary-wkr-2 \
    --public-addr $WORKER2_PUBLIC_IP \
    --worker-auth-key $WORKER_AUTH_KEY \
    --controller $CONTROLLER_PRIVATE_IP