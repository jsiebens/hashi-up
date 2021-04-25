resource "digitalocean_droplet" "db" {
  image     = "docker-20-04"
  name      = "boundary-db"
  region    = var.region
  size      = "s-1vcpu-1gb"
  vpc_uuid  = var.vpc_uuid
  ssh_keys  = [var.ssh_key_id]
  user_data = file("./scripts/setup_db.sh")
}

resource "digitalocean_droplet" "controller" {
  image    = "ubuntu-20-04-x64"
  name     = "boundary-controller"
  region   = var.region
  size     = "s-1vcpu-1gb"
  vpc_uuid = var.vpc_uuid
  ssh_keys = [var.ssh_key_id]
  user_data = file("./scripts/setup.sh")
}

resource "digitalocean_droplet" "worker1" {
  image    = "ubuntu-20-04-x64"
  name     = "boundary-worker-01"
  region   = var.region
  size     = "s-1vcpu-1gb"
  vpc_uuid = var.vpc_uuid
  ssh_keys = [var.ssh_key_id]
  user_data = file("./scripts/setup_db.sh")
}

resource "digitalocean_droplet" "worker2" {
  image    = "ubuntu-20-04-x64"
  name     = "boundary-worker-02"
  region   = var.region
  size     = "s-1vcpu-1gb"
  vpc_uuid = var.vpc_uuid
  ssh_keys = [var.ssh_key_id]
  user_data = file("./scripts/setup_db.sh")
}

output "db_public_ip" {
  value = digitalocean_droplet.db.ipv4_address
}

output "db_private_ip" {
  value = digitalocean_droplet.db.ipv4_address_private
}

output "controller_public_ip" {
  value = digitalocean_droplet.controller.ipv4_address
}

output "controller_private_ip" {
  value = digitalocean_droplet.controller.ipv4_address_private
}

output "worker1_public_ip" {
  value = digitalocean_droplet.worker1.ipv4_address
}

output "worker1_private_ip" {
  value = digitalocean_droplet.worker1.ipv4_address_private
}

output "worker2_public_ip" {
  value = digitalocean_droplet.worker2.ipv4_address
}

output "worker2_private_ip" {
  value = digitalocean_droplet.worker2.ipv4_address_private
}