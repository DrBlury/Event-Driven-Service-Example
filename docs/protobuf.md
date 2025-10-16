# Protobuf and Domain Models

This project uses Protobuf for defining domain models. Below are the reasons and benefits:

## Why Protobuf?

1. **Centralized Validation Logic**:
   - Protobuf ensures that the data structure is validated centrally, reducing redundancy.

2. **Language-Agnostic Domain Models**:
   - Protobuf files can be used across different programming languages, ensuring consistency.

3. **Ease of Communication**:
   - Protobuf is generally widely used for communication between services, making it easier to integrate with other systems in the future. It offers wide support of gRPC and messaging systems.

4. **Overhead Reduction**:
   - The way that Protobuf is used in this project reduces the overhead of maintaining separate domain models for different services. It also allows you to skip implementing validation logic manually in your go code. Instead all validation and resulting errors will be generated automatically and are in one common format. No implementation difference.

## Generating Protobuf Files

- Use the following command to generate the protobuf files:

  ```bash
  task gen-buf
  ```
