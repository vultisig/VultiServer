# VultiSigner Demo Dashboard

This project is a demo dashboard built with React.js for interacting with VultiSigner. It facilitates the creation of new vaults and allows two parties to join and manage them.

## Features

- **Define Vault Name:** Users can define a name for the vault they wish to create.
- **Generate QR Code:** Calls an endpoint in VultiSigner to generate a new QR code for client interaction.
- **Wait for Parties:** The demo app waits for parties to join and proceed with key generation.
- **Select Parties:** Users can select the parties involved in the vault creation and management process.
- **Generate Keys:** Calls an endpoint in relayer and check for the key generation status.
- 
Please note that VultiSigner itself does not generate keys; this demo is solely for demonstrating the vault creation and management process between different parties.

## Configuration and Environment Setup

To update configuration or environment variables, run the following command to regenerate the HTML, CSS, and JS files for the demo app:

```bash
make generate-demo
```
