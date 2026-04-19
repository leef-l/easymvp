import { mkdir } from "node:fs/promises";
import path from "node:path";
import { spawn } from "node:child_process";

const desktopRoot = path.resolve(import.meta.dirname, "..");
const repoRoot = path.resolve(desktopRoot, "..", "..");
const coreRoot = path.join(repoRoot, "apps", "core");
const outputDir = path.join(desktopRoot, "resources", "bin");
const isWindows = process.platform === "win32";
const binaryName = isWindows ? "easymvp-core.exe" : "easymvp-core";
const outputPath = path.join(outputDir, binaryName);

const env = {
  ...process.env,
  CGO_ENABLED: "0",
};

if (!env.GOOS) {
  env.GOOS = isWindows ? "windows" : process.platform;
}

if (!env.GOARCH) {
  env.GOARCH = process.arch === "x64" ? "amd64" : process.arch;
}

await mkdir(outputDir, { recursive: true });

await new Promise((resolve, reject) => {
  const child = spawn(
    "go",
    ["build", "-o", outputPath, "./main.go"],
    {
      cwd: coreRoot,
      env,
      stdio: "inherit",
    },
  );

  child.on("exit", (code) => {
    if (code === 0) {
      resolve();
      return;
    }
    reject(new Error(`go build failed with exit code ${code ?? -1}`));
  });

  child.on("error", reject);
});
