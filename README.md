## SpartaHoneycomb

Seamlessly publish your Sparta logrus output to honeycomb.io 🐝

### Instructions

This repository uses [SpartaVault](https://github.com/mweagle/SpartaVault)
to manage your honeycomb.io [WriteKey](https://honeycomb.io/docs/reference/api/).
The encrypted version is included in the source code and uses KMS to lazily
decrypt the key on the first lambda execution.

To successfully provision this application:
  1. Ensure you have a [KMS Key](https://aws.amazon.com/kms/)
  1. Create a [Honeycomb.io](https://honeycomb.io) account and make note of your *WriteKey*
  1. Use the `SpartaVault` command line [tool](https://github.com/mweagle/SpartaVault#usage)
    to encrypt your key as in: 
      ```
      SpartaVault encrypt \
        --key {YOUR_KMS_KEY_ARN} \
        --value "{YOUR_HONEYCOMB_IO_WRITE_KEY}" \
        --name "HoneycombWriteKey"
      ```
  1. Update the `HoneycombWriteKey` in _main.go_ with the CLI output
  1. Optionally update the default Honeycomb Dataset name (`LambdaDataset`)
  1. Provision!
