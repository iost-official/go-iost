# IOST - A High-Performance, Developer-Friendly Blockchain

IOST is a smart contract platform designed with a strong emphasis on performance and developer experience. It provides a robust infrastructure for building complex blockchain applications with ease and efficiency.

# Features

1. **V8 JavaScript Engine**: The V8 JavaScript engine is integrated directly into the blockchain, enabling developers to write smart contracts using JavaScript. This brings about a world of benefits as JavaScript is a universally known language, making the development process easier and quicker. Furthermore, the V8 engine, being Google Chrome's open-source JavaScript engine, is known for its speed, which translates into extremely efficient smart contract execution.

2. **High Scalability**: IOST blockchain delivers high throughput with thousands of transactions per second (TPS), making it suitable for large-scale applications. While maintaining this high scalability, it also employs a more decentralized consensus mechanism than DPoS (Delegated Proof of Stake), ensuring a balance between performance and decentralization.

3. **Fast Block Times and Finality**: With a 0.5-second block time and 0.5-minute finality, developers can create responsive, real-time applications. This rapid finality ensures transactions are confirmed quickly, in less than a minute. This is a highly desirable feature for developers as it improves user experience by providing swift, secure transaction confirmations.

4. **Free Transactions**: IOST offers free transactions. Users can stake coins to acquire gas, eliminating the need for transaction fees and making the platform more user-friendly.

# Development

### Environment Requirements

- OS: Ubuntu 22.04 or later  
- Go: 1.20 or later

Please note that the IOST node utilizes the CGO V8 JavaScript engine and currently only supports the x64 architecture.

### Deployment

- Build a local binary: Run `make build`
- Start a local development network: Run `make debug`
- Build a Docker image: Run `make image`

For comprehensive documentation, please visit: [IOST Developer](https://developers.iost.io)

Join our [tech community on Telegram](https://t.me/iostdev) for discussions, updates, and support.

Happy hacking! We look forward to seeing what you build on the IOST platform.

