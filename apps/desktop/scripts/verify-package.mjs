import { access } from "node:fs/promises";
import { spawn } from "node:child_process";
import path from "node:path";

const desktopRoot = path.resolve(import.meta.dirname, "..");
const smokeTimeoutMs = 30000;
const packageTargets = [
  {
    releaseRoot: path.join(desktopRoot, "release", "linux-unpacked"),
    binaryPath: path.join(
      desktopRoot,
      "release",
      "linux-unpacked",
      "resources",
      "bin",
      "easymvp-core",
    ),
    executablePath: path.join(
      desktopRoot,
      "release",
      "linux-unpacked",
      "easymvp-desktop",
    ),
  },
  {
    releaseRoot: path.join(desktopRoot, "release", "win-unpacked"),
    binaryPath: path.join(
      desktopRoot,
      "release",
      "win-unpacked",
      "resources",
      "bin",
      "easymvp-core.exe",
    ),
    executablePath: path.join(
      desktopRoot,
      "release",
      "win-unpacked",
      "EasyMVP.exe",
    ),
  },
];

const packageTarget = await resolvePackageTarget();
const { releaseRoot, binaryPath, executablePath } = packageTarget;

console.log(`verified desktop package at ${releaseRoot}`);
console.log(`verified bundled core binary at ${binaryPath}`);

await runSmokeTest({
  command: executablePath,
  args: ["--smoke-test"],
});

async function resolvePackageTarget() {
  for (const target of packageTargets) {
    if (
      (await pathExists(target.releaseRoot))
      && (await pathExists(target.binaryPath))
      && (await pathExists(target.executablePath))
    ) {
      return target;
    }
  }
  throw new Error("no packaged desktop directory build found under release/");
}

async function pathExists(targetPath) {
  try {
    await access(targetPath);
    return true;
  } catch {
    return false;
  }
}

async function runSmokeTest({ command, args }) {
  await new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: releaseRoot,
      env: {
        ...process.env,
        EASYMVP_MANAGE_CORE: "1",
        ELECTRON_DISABLE_SANDBOX: "1",
      },
      stdio: ["ignore", "pipe", "pipe"],
    });

    let stdout = "";
    let stderr = "";
    const timer = setTimeout(() => {
      child.kill("SIGTERM");
      reject(
        new Error(
          `desktop smoke test timed out after ${smokeTimeoutMs}ms\nstdout:\n${stdout}\nstderr:\n${stderr}`,
        ),
      );
    }, smokeTimeoutMs);

    child.stdout.on("data", (chunk) => {
      stdout += chunk.toString();
    });
    child.stderr.on("data", (chunk) => {
      stderr += chunk.toString();
    });

    child.on("error", (error) => {
      clearTimeout(timer);
      reject(error);
    });

    child.on("exit", (code) => {
      clearTimeout(timer);
      if (code === 0) {
        resolve(undefined);
        return;
      }
      reject(
        new Error(
          `desktop smoke test failed with exit code ${code}\nstdout:\n${stdout}\nstderr:\n${stderr}`,
        ),
      );
    });
  });

  console.log(`verified packaged desktop smoke at ${executablePath}`);
}
