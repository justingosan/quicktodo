# Publishing QuickTodo to Homebrew

## Prerequisites

1. Create a GitHub release at https://github.com/justingosan/quicktodo/releases
2. Upload the release archives from `build/releases/`

## Steps to Publish Your Tap

### 1. Create the tap repository on GitHub

1. Create a new repository named `homebrew-quicktodo` on GitHub
   - **IMPORTANT**: The repository name MUST start with `homebrew-`
   - Make it public

### 2. Push the tap files

```bash
cd homebrew-quicktodo
git init
git add .
git commit -m "Initial homebrew tap for QuickTodo v1.0.0"
git branch -M main
git remote add origin https://github.com/justingosan/homebrew-quicktodo.git
git push -u origin main
```

### 3. Test the tap locally

```bash
# Add your tap
brew tap justingosan/quicktodo

# Install quicktodo
brew install quicktodo

# Test it works
quicktodo version
```

### 4. Users can now install with:

```bash
brew tap justingosan/quicktodo
brew install quicktodo
```

Or in a single command:

```bash
brew install justingosan/quicktodo/quicktodo
```

## Updating for New Releases

1. Build new release archives:
   ```bash
   git tag vX.Y.Z
   make release
   ```

2. Upload to GitHub releases

3. Update the formula:
   - Update version number
   - Update URLs to point to new release
   - Update SHA256 hashes using `shasum -a 256 <file>`

4. Commit and push to homebrew-quicktodo repository

## Alternative: Submit to homebrew-core

Once your project has:
- 50+ GitHub stars
- Stable releases
- Active maintenance

You can submit to the official homebrew-core repository for wider distribution.