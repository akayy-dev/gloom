## Features

Gloom is an open-source financial terminal designed to provide users with powerful tools for financial analysis and data visualization. Key features include:

- **Real-time Market Data**: Access live market data for stocks, commodities, and more.
- **News Aggregation**: Utilize RSS to read all of your financial news in one place
- **Open Source**: Fully open-source and customizable to suit your needs.

## Environment Variables
To run Gloom, you need to set the following environment variables:

| Variable Name | Description                                            |
|---------------|--------------------------------------------------------|
| `TIINGO_KEY`  | API Key for accessing stock data                       |
| `GEMINI_KEY`  | API Key for using Google Gemini to web scrape articles |
| `SSH_HOST`    | URL to expose the SSH server                           |


Make sure to set these variables in your environment before starting the application.

## Getting Started

1. Clone the repository:
    ```bash
    git clone https://github.com/akayy-dev/gloom.git
    ```
2. Install dependencies:
    ```bash
    cd gloom
    go mod tidy
    ```
3. Set the required environment variables.
4. Start the application:
    ```bash
    go run ./cmd/ui
    ```

# Configuration
You can configure the behavior of the app using a `config.json` file. Gloom searches in `$HOME/.config/gloom/` for the config file. A default config file can be found [here](./internal/shared/config/default.json)
