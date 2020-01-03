provider "azurerm" {
  version = "~> 1.39"
}

resource "azurerm_resource_group" "hitpoints" {
  name     = "hitpoints"
  location = var.resource_group_location
}

resource "azurerm_storage_account" "hitpoints" {
  name                     = var.prefix
  resource_group_name      = azurerm_resource_group.hitpoints.name
  location                 = azurerm_resource_group.hitpoints.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  tags                     = var.tags
}

resource "azurerm_storage_container" "hitpoints" {
  name                  = "outputs"
  storage_account_name  = azurerm_storage_account.hitpoints.name
  container_access_type = "blob"
}

resource "azurerm_virtual_network" "hitpoints" {
  name                = "${var.prefix}-network"
  address_space       = ["10.0.0.0/16"]
  resource_group_name = azurerm_resource_group.hitpoints.name
  location            = azurerm_resource_group.hitpoints.location
}

resource "azurerm_public_ip" "hitpoints" {
  name                    = "${var.prefix}-ip"
  resource_group_name     = azurerm_resource_group.hitpoints.name
  location                = azurerm_resource_group.hitpoints.location
  allocation_method       = "Static"
  idle_timeout_in_minutes = 30
  tags                    = var.tags
}

resource "azurerm_network_security_group" "hitpoints" {
  name                = "${var.prefix}-nsg"
  resource_group_name = azurerm_resource_group.hitpoints.name
  location            = azurerm_resource_group.hitpoints.location

  security_rule {
    name                       = "AllowSSHInbound"
    description                = "Allow inbound SSH traffic on port 22"
    priority                   = 400
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = 22
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "AllowHTTPInbound"
    description                = "Allow inbound traffic on port 80"
    priority                   = 500
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = 80
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "AllowHTTPSInbound"
    description                = "Allow inbound traffic on port 443"
    priority                   = 501
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = 443
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "DenyAllInbound"
    description                = "Deny all inbound traffic"
    priority                   = 550
    direction                  = "Inbound"
    access                     = "Deny"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "AllowAllOutbound"
    description                = "Allow all outbound traffic"
    priority                   = 551
    direction                  = "Outbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

resource "azurerm_subnet" "external" {
  name                 = "${var.prefix}external"
  resource_group_name  = azurerm_resource_group.hitpoints.name
  virtual_network_name = azurerm_virtual_network.hitpoints.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurerm_subnet_network_security_group_association" "external" {
  subnet_id                 = azurerm_subnet.external.id
  network_security_group_id = azurerm_network_security_group.hitpoints.id
}

resource "azurerm_network_interface" "hitpoints" {
  name                      = "${var.prefix}-nic"
  resource_group_name       = azurerm_resource_group.hitpoints.name
  location                  = azurerm_resource_group.hitpoints.location
  network_security_group_id = azurerm_network_security_group.hitpoints.id

  ip_configuration {
    name                          = "${var.prefix}ipconfig"
    subnet_id                     = azurerm_subnet.external.id
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = azurerm_public_ip.hitpoints.id
  }
}

resource "azurerm_virtual_machine" "hitpoints" {
  name                          = "${var.prefix}-vm"
  resource_group_name           = azurerm_resource_group.hitpoints.name
  location                      = azurerm_resource_group.hitpoints.location
  network_interface_ids         = [azurerm_network_interface.hitpoints.id]
  vm_size                       = var.vm_size
  delete_os_disk_on_termination = true
  tags                          = var.tags

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "${var.prefix}-disk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
    disk_size_gb      = 30
  }

  os_profile {
    computer_name  = "${var.prefix}-vm"
    admin_username = "ubuntu"
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      key_data = var.ssh_pub_key
      path     = "/home/ubuntu/.ssh/authorized_keys"
    }
  }

  connection {
    type        = "ssh"
    user        = "ubuntu"
    host        = azurerm_public_ip.hitpoints.ip_address
    private_key = file(var.ssh_priv_key_path)
  }

  provisioner "file" {
    content = templatefile("${path.module}/../files/hitpoints.service", {
      ssl          = var.ssl
      domain       = var.domain
      command      = "azure"
      command_args = "--account-name ${azurerm_storage_account.hitpoints.name} --account-key ${azurerm_storage_account.hitpoints.primary_access_key} --container ${azurerm_storage_container.hitpoints.name}"
    })
    destination = "/home/ubuntu/hitpoints.service"
  }

  provisioner "file" {
    source      = "${path.module}/../../release/hitpoints-linux-amd64/hitpoints"
    destination = "/home/ubuntu/hitpoints"
  }

  provisioner "remote-exec" {
    script = "${path.module}/../files/setup-hitpoints.sh"
  }
}
