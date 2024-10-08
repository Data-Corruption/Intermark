# Getting Started

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Setup](#setup)
- [Running the Application](#running-the-application)
- [Managing Site Content](#managing-site-content)
  - [Editing Site Layout](#editing-site-layout)
  - [Automating Content Updates](#automating-content-updates)

## Prerequisites

*Note! - Currently I'm only supporting linux officially. If enough people request it, I'll add official support for windows.*

Ensure you have the following installed:

- **git and unzip**: Install via your package manager (e.g., `sudo apt-get install git unzip`)
- **Go**: [Download and Install Go](https://go.dev/doc/install)
- **Node.js**: [Download and Install Node.js](https://nodejs.org/en/download/package-manager)

## Installation

1. **Clone The Repository**: `git clone https://github.com/Data-Corruption/intermark.git`
2. **Enter The Project Root**: `cd intermark`
3. **Install Dependencies**: `npm install`
4. **Build the Project**: `./build.sh`
5. **Generate Configuration File**: `./bin/intermark-linux-amd64`

## Setup

1. **Create a Deploy Key**:

   Generate a new SSH key specifically for your Intermark site:

   ```bash
   ssh-keygen -t ed25519 -C "comment for key, machine, email, whatever" -f ~/.ssh/id_ed25519_intermark
   ```

2. **Configure SSH for GitHub Access**:

   Add 'github-intermark' to your SSH config for the site to use when reading your content repo. Edit `~/.ssh/config` and add:

   ```conf
   Host github-intermark
       HostName github.com
       User git
       IdentityFile ~/.ssh/id_ed25519_intermark
   ```

   This allows having multiple github ssh keys on the same system and won't interfere if you already have a github ssh key in use on the machine.

   If you don't want to use 'github-intermark' you can change it but you'll also need to update the config you'll see later.

3. **Retrieve the Public Key**:

   Copy your public key for later use:

   ```bash
   cat ~/.ssh/id_ed25519_intermark.pub
   ```

4. **Create a Content Repository on GitHub**:

   - Go to GitHub and create a new repository for your content.

5. **Add the Deploy Key to the Repository**:

    - In your content repository, go to **Settings** > **Deploy keys**.
    - Click **Add deploy key**, give it a title, and paste the public key from earlier.

6. **Enable GitHub Actions Permissions**:

   - Navigate to **Settings** > **Actions** > **General**.
   - Under **Workflow permissions**, select **Read and write permissions**.

7. **Add Workflow to Content Repository**:

   - Copy the `.github` directory from the `content_template/` folder in this repository.
   - Paste it in the root of your content repository.

8. **Run the Workflow**:

    - Push changes to trigger the workflow, or run it manually via the **Actions** tab on GitHub. The workflow adds an ID to the top of all `.md` files, which is used to track content and prevent dead links.

9. **Update Application Configuration**:

    - Copy the SSH link for the content repository (e.g., `git@github.com:username/content-repo.git`).
    - Update the configuration file generated earlier, setting **content_repo** > **url** to your link.

## Running the Application

Start the application:

```bash
./bin/intermark-linux-amd64
```

## Managing Site Content

### Editing Site Layout

1. **Access the Edit GUI**:

   - Upon running the application, a link to the edit GUI is provided.
   - Default password is empty. You can set the password in the config file. You'll need to restart the app if you edit the config while it's running.

2. **Create Pages and Assigning Content**:

   - Use the Edit GUI to create new pages
   - Click the Update Content button to fetch the latest state of your content repo.
   - Assign content from your content repository to these pages.

   <!-- TODO: Add gif with captions that demonstrates the above steps -->

3. **Managing Assets**:

   - To use other files(images, scripts, etc) in your markdown, commit the files to your content repository at `./assets`. You can then use them in your `.md` files like so:

      ```markdown
      ![Example Image](/assets/example_image.png)
      <script type="text/javascript" src="/assets/example_script.js" id="example_script">
      ```

      You can edit the asset folder path via the config as well.

   <!-- TODO: Add gif with captions that demonstrates the above steps -->

4. **Deleting Content**:

   If you try to delete a .md file directly from the content repository, the GitHub Actions workflow will error out. This safeguard prevents the accidental creation of dead links on the site.

   To properly delete a file:

   - Remove the `.md` file from your content repository.
   - In the **same commit**, delete the corresponding entry from `.github/ids.json`.

   <!-- TODO: Add gif with captions that demonstrates the above steps -->

   This functions as a delete confirmation. By updating ids.json alongside the file deletion, you inform the workflow of your intent, allowing it to process the change without errors.

### Automating Content Updates

To automatically update content when changes are pushed to the content repository:

1. **Expose Your Application to the Internet**:

   - Ensure your server is accessible from the internet at `http://your-server-address:port`. e.g. forward the port, update firewall setting if needed.

2. **Set Environment Variables in GitHub Actions**:

   - In your content repository, go to **Settings** > **Secrets and variables** > **Actions**.
   - Add a new **Repository Variable** named `SERVER_ADDRESS` with the value of your server's address (e.g., `http://your-server-address:port`).
   - Generate a random alphanumeric secret (e.g., `openssl rand -hex 16`).
   - Add a new **Repository Secret** named `UPDATE_TOKEN` with the random secret you just generated.

3. **Update The Apps Config**:

   - Set the **content_repo** > **update_token** in your app's config file to match the UPDATE_TOKEN you set in GitHub.
   - Restart your app if it was running.

Now, when you push updates to the content repository, GitHub Actions will notify your server to update its content. This ensures your pages automagically reflect any edits made.

<!-- TODO: Add gif with captions that demonstrates the above statement -->
