-- Migration down: drop device_bootstrap_tokens and authorized_devices tables

DROP TABLE IF EXISTS device_bootstrap_tokens;
DROP TABLE IF EXISTS authorized_devices;
