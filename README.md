# Project Name

## Overview

This project is a [brief description of your application]. It is built using [programming language/framework] and provides [key features or functionalities].

## Prerequisites

- [Programming language] version [version]
- [Build tool] version [version]
- AWS CLI version [version]
- [Any other dependencies]

## Building the Application

1. Clone the repository:

   ```sh
   git clone https://github.com/bernhardrode/udp-mqtt-bridge.git
   cd udp-mqtt-bridge
   ```

2. Install dependencies:

   ```sh
   go mod download
   ```

3. Build the application:

   ```sh
   go build -o udd_mqtt_bridge ./cmd/main.go
   ```

## Creating a Configuration File

1. Create a configuration file named `config.yaml`:

   ```sh
   touch config.yaml
   ```

2. Edit the `config.yaml` file and add the necessary configuration parameters for your application.

> **Note:** You can use the `configs/config.sample.yaml` file in the repository as a template.

## Downloading Certificates

1. Download the certificate files `cert.pem`, `public.key`, and `private.key` that were generated in the previous steps.

2. Place the downloaded certificate files in the same directory as the `udd_mqtt_bridge` executable.

## Creating an IoT Thing and Downloading Certificates Using AWS CLI

### Step 1: Create an IoT Thing

1. Create a new IoT thing:

   ```sh
   aws iot create-thing --thing-name your-thing-name
   ```

   Replace [`your-thing-name`](command:_github.copilot.openSymbolFromReferences?%5B%22your-thing-name%22%2C%5B%7B%22uri%22%3A%7B%22%24mid%22%3A1%2C%22fsPath%22%3A%22%2Fhome%2Frbo2abt%2Fdev%2Fbosch-engineering%2Fudp_mqtt_bridge%2FREADME.md%22%2C%22external%22%3A%22file%3A%2F%2F%2Fhome%2Frbo2abt%2Fdev%2Fbosch-engineering%2Fudp_mqtt_bridge%2FREADME.md%22%2C%22path%22%3A%22%2Fhome%2Frbo2abt%2Fdev%2Fbosch-engineering%2Fudp_mqtt_bridge%2FREADME.md%22%2C%22scheme%22%3A%22file%22%7D%2C%22pos%22%3A%7B%22line%22%3A40%2C%22character%22%3A26%7D%7D%5D%5D "Go to definition") with the desired name for your IoT thing.

### Step 2: Get the AWS IoT Endpoint

1. Retrieve the AWS IoT endpoint:

   ```sh
   aws iot describe-endpoint --endpoint-type iot:Data-ATS
   ```

   This command will return the endpoint URL that your IoT device will use to communicate with AWS IoT.

### Step 3: Create and Download Certificates

1. Create a new certificate and keys:

   ```sh
   aws iot create-keys-and-certificate --set-as-active --certificate-pem-outfile cert.pem --public-key-outfile public.key --private-key-outfile private.key
   ```

   This command will generate a certificate and keys, and save them to `cert.pem`, `public.key`, and `private.key` respectively.

2. Attach the certificate to your IoT thing:

   ```sh
   aws iot attach-thing-principal --thing-name your-thing-name --principal arn:aws:iot:region:account-id:cert/certificate-id
   ```

   Replace [`your-thing-name`](command:_github.copilot.openSymbolFromReferences?%5B%22your-thing-name%22%2C%5B%7B%22uri%22%3A%7B%22%24mid%22%3A1%2C%22fsPath%22%3A%22%2Fhome%2Frbo2abt%2Fdev%2Fbosch-engineering%2Fudp_mqtt_bridge%2FREADME.md%22%2C%22external%22%3A%22file%3A%2F%2F%2Fhome%2Frbo2abt%2Fdev%2Fbosch-engineering%2Fudp_mqtt_bridge%2FREADME.md%22%2C%22path%22%3A%22%2Fhome%2Frbo2abt%2Fdev%2Fbosch-engineering%2Fudp_mqtt_bridge%2FREADME.md%22%2C%22scheme%22%3A%22file%22%7D%2C%22pos%22%3A%7B%22line%22%3A40%2C%22character%22%3A26%7D%7D%5D%5D "Go to definition") with the name of your IoT thing, and `arn:aws:iot:region:account-id:cert/certificate-id` with the ARN of the certificate created in the previous step.

### Step 4: Attach a Policy to the Certificate

1. Create an IoT policy (if you don't have one already):

   ```sh
   aws iot create-policy --policy-name your-policy-name --policy-document file://policy.json
   ```

   Replace [`your-policy-name`](command:_github.copilot.openSymbolFromReferences?%5B%22your-policy-name%22%2C%5B%7B%22uri%22%3A%7B%22%24mid%22%3A1%2C%22fsPath%22%3A%22%2Fhome%2Frbo2abt%2Fdev%2Fbosch-engineering%2Fudp_mqtt_bridge%2FREADME.md%22%2C%22external%22%3A%22file%3A%2F%2F%2Fhome%2Frbo2abt%2Fdev%2Fbosch-engineering%2Fudp_mqtt_bridge%2FREADME.md%22%2C%22path%22%3A%22%2Fhome%2Frbo2abt%2Fdev%2Fbosch-engineering%2Fudp_mqtt_bridge%2FREADME.md%22%2C%22scheme%22%3A%22file%22%7D%2C%22pos%22%3A%7B%22line%22%3A40%2C%22character%22%3A26%7D%7D%5D%5D "Go to definition") with the desired name for your policy, and ensure `policy.json` contains the appropriate policy document.

2. Attach the policy to the certificate:

   ```sh
   aws iot attach-policy --policy-name your-policy-name --target arn:aws:iot:region:account-id:cert/certificate-id
   ```

   Replace [`your-policy-name`](command:_github.copilot.openSymbolFromReferences?%5B%22your-policy-name%22%2C%5B%7B%22uri%22%3A%7B%22%24mid%22%3A1%2C%22fsPath%22%3A%22%2Fhome%2Frbo2abt%2Fdev%2Fbosch-engineering%2Fudp_mqtt_bridge%2FREADME.md%22%2C%22external%22%3A%22file%3A%2F%2F%2Fhome%2Frbo2abt%2Fdev%2Fbosch-engineering%2Fudp_mqtt_bridge%2FREADME.md%22%2C%22path%22%3A%22%2Fhome%2Frbo2abt%2Fdev%2Fbosch-engineering%2Fudp_mqtt_bridge%2FREADME.md%22%2C%22scheme%22%3A%22file%22%7D%2C%22pos%22%3A%7B%22line%22%3A40%2C%22character%22%3A26%7D%7D%5D%5D "Go to definition") with the name of your policy, and `arn:aws:iot:region:account-id:cert/certificate-id` with the ARN of the certificate.

### Step 5: Download the Root CA Certificate

1. Download the Amazon Root CA certificate:

   ```sh
   wget https://www.amazontrust.com/repository/AmazonRootCA1.pem
   ```

## Running the Application

1. Start the application:

```sh
./udd_mqtt_bridge
```
