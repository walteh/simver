# SimVer - Semantic Versioning Simplified 🔄

## 🌟 Overview

Forget about the headache of manual versioning. SimVer is here to automate and simplify semantic versioning in your Git workflow. By managing version tags directly linked to your commit activities, SimVer ensures a linear, predictable version history with zero manual overhead. Perfect for teams and developers looking for a set-it-and-forget-it solution, SimVer streamlines your development process, allowing you to focus on what truly matters - building great software.

## 🎯 Target Audience

SimVer is the go-to solution for developers and teams who:

-   Are exhausted by manual versioning complexities.
-   Desire a consistent, traceable version history.
-   Need seamless integration with CI/CD tools like GitHub Actions.

## ✨ Features

-   **Automated Version Tags:** Automatically generates semantic version tags for all commits and pull requests.
-   **Clear Versioning Rules:** Simplifies decision-making with straightforward rules for main and side branches.
-   **Linear History Dependence:** Ensures all version increments are predictable and orderly.

## 📘 How It Works

### Versioning Logic Simplified

-   **Direct Pushes to Main:** Trigger minor version updates automatically.
-   **Pull Requests:**
    -   To Main: Trigger minor updates.
    -   To Side Branches: Apply patch updates.
    -   Multiple Concurrent PRs: Ensure unique, sequential versioning without conflicts.
-   **Merging:** Versions merge seamlessly, with the target branch adopting the version from the merged branch or PR.

### Tagging Made Easy

-   **Main Branch:** Receives clean, full release versions.
-   **Side Branches and PRs:** Get detailed pre-release tags to avoid confusion and maintain clarity.

### Real-world Examples

-   Direct push to `main` changes `v1.0.0` to `v1.1.0`.
-   A PR to `main` updates from `v1.0.0` to `v1.1.0`.
-   A PR to a side branch updates from `v1.1.0` to `v1.1.1`.
-   Handling multiple PRs ensures sequential updates without overlap.

## 🚀 Getting Started

### Usage

Easily integrate SimVer into your GitHub Actions with this setup:

```yaml
name: simver
permissions: { id-token: write, contents: write, pull-requests: read }
on:
    workflow_dispatch:
    push:
        branches: [main]
    pull_request:
        types: [opened, synchronize, reopened, closed]
jobs:
    simver:
        runs-on: ubuntu-latest
        steps:
            - uses: walteh/simver/cmd/gha-simver@v0
              with:
                  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## ⚠️ Current Limitations & 🛠 Future Fixes

-   **Junk Tags Cleanup:** Upcoming feature to clear temporary tags automatically. (#13)
-   **Force Push Handling:** We're improving how version recalculations handle force pushes to maintain accurate histories. (#6)

## 🤝 Contributing

Contributions are welcome! Fork, modify, and submit a pull request.

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.
