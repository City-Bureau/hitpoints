output "address" {
  value = "${aws_eip.hitpoints.public_ip}"
}
