package main

// DHCP Server
const DHCP_SERVER_ADDR = "172.10.0.1"
const DHCP_SERVER_LEASE_START_ADDR = "172.10.0.2"
const DHCP_LEASE_DURATION = 10
const DHCP_LEASE_COUNT = 10
const DHCP_LEASE_RANGE = 50

// TFTP Server
const TFTP_SERVER_ADDR = "172.20.0.1"
const PXELINUX_LOADER = "undionly.kpxe"

// DNS Server
// TODO: Implement DNS Server within xserver
const DNS_SERVER_ADDR = "172.10.0.1"
