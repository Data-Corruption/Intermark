# <img src="/docs/assets/logo-darkmode.png" height="40" align="left" alt="logo">Intermark

- [Getting Started](/docs/getting-started.md)
- [Styling](/docs/styling.md)
- [Troubleshooting](/docs/troubleshooting.md)
- [Contributing](/docs/contributing.md)
- [License](/LICENSE.md)

**Intermark** is a hybrid Content Management System / Static Site Generator, perfect for collaborative blogs or educational sites. In a nutshell... point this app at a github repo containing `.md` files, drop a workflow into it, then the app will help you easily turn that repo into a beautiful website and keep it updated.

This system aims to achieve several objectives simultaneously:

- **Flexible Sidebar**: Allow interlacing files, folders, and dividers.
- **Content in a Repository**: Utilize a GitHub repository for content and assets, allowing for writing to be more collaborative and manageable.
- **Full HTML Support**: Optionally use html in the markdown with [tailwindcss](https://tailwindcss.com/) and [daisyui](https://daisyui.com/) support.
- **Robust Linking Mechanism**:
  - **Safe Moving and Renaming**: Move or rename `.md` files without risking dead links.
  - **Deletion Safeguard**: Require a secondary action when deleting `.md` files to prevents accidental deletions that could result in broken links.
- **Robust Synchronization System**: Automatically synchronize content updates from the content repository to the site, ensuring pages reflect the latest versions.
- **Use Golang**: I like Go rawr xD uwu owo <3.

*Note - Currently I'm only supporting linux officially. If enough people request it, I'll add official support for windows.*

<!--

Repo name / desc: Intermark - hybrid CMS/SSG, perfect for educational sites or collaborative blogs.

TODO: final pass / sanity run. Test everything. Make gifs for readme

-->
