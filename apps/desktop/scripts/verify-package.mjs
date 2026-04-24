import { access } from "node:fs/promises";
import { spawn } from "node:child_process";
import net from "node:net";
import path from "node:path";

const desktopRoot = path.resolve(import.meta.dirname, "..");
const smokeTimeoutMs = 45000;
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

const smokePort = await allocateSmokePort();
const smokeBaseUrl = `http://127.0.0.1:${smokePort}`;

await runSmokeTest({
  command: executablePath,
  args: ["--smoke-test"],
  baseUrl: smokeBaseUrl,
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

async function allocateSmokePort() {
  return await new Promise((resolve, reject) => {
    const server = net.createServer();
    server.once("error", reject);
    server.listen(0, "127.0.0.1", () => {
      const address = server.address();
      server.close(() => {
        if (address && typeof address === "object") {
          resolve(address.port);
          return;
        }
        reject(new Error("failed to allocate a smoke test port"));
      });
    });
  });
}

async function runSmokeTest({ command, args, baseUrl }) {
  const { stdout, stderr } = await new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: releaseRoot,
      env: {
        ...process.env,
        EASYMVP_MANAGE_CORE: "1",
        EASYMVP_CORE_BASE_URL: baseUrl,
        ELECTRON_DISABLE_SANDBOX: "1",
        EASYMVP_DESKTOP_SMOKE: "1",
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
        resolve({ stdout, stderr });
        return;
      }
      reject(
        new Error(
          `desktop smoke test failed with exit code ${code}\nstdout:\n${stdout}\nstderr:\n${stderr}`,
        ),
      );
    });
  });

  verifySmokeOutput({ stdout, stderr, baseUrl });

  console.log(`verified packaged desktop smoke at ${executablePath}`);
  console.log(`verified packaged desktop managed core smoke at ${baseUrl}`);
}

function verifySmokeOutput({ stdout, stderr, baseUrl }) {
  const lines = `${stdout}\n${stderr}`
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line.startsWith("EASYMVP_SMOKE "));
  const entries = lines.map(parseSmokeLine);

  const lifecycle = entries.find((entry) => entry.managed !== undefined);
  assertSmoke(
    lifecycle?.managed === "true",
    "smoke output did not prove managed=true",
    stdout,
    stderr,
  );
  assertSmoke(
    lifecycle?.phase === "ready",
    "smoke output did not prove phase=ready",
    stdout,
    stderr,
  );
  assertSmoke(
    lifecycle?.baseUrl === baseUrl,
    `smoke output did not report expected baseUrl=${baseUrl}`,
    stdout,
    stderr,
  );
  assertSmoke(
    Number(lifecycle?.corePid || 0) > 0,
    "smoke output did not report a managed core pid",
    stdout,
    stderr,
  );

  const health = entries.find((entry) => entry.healthOk !== undefined);
  const httpStatus = Number(health?.httpStatus || 0);
  assertSmoke(
    health?.healthOk === "true"
      && httpStatus >= 200
      && httpStatus < 300
      && health?.endpoint === "/api/v3/system/healthz",
    "smoke output did not prove real /api/v3/system/healthz health OK",
    stdout,
    stderr,
  );

  const renderer = entries.find((entry) => entry.rendererMounted !== undefined);
  assertSmoke(
    renderer?.rendererMounted === "true" && Number(renderer?.rootChildCount || 0) > 0,
    "smoke output did not prove renderer mounted",
    stdout,
    stderr,
  );

  const cleanup = entries.find((entry) => entry.cleanup !== undefined);
  assertSmoke(
    cleanup?.cleanup === "stopped",
    "smoke output did not prove managed core cleanup",
    stdout,
    stderr,
  );
}

function parseSmokeLine(line) {
  const fields = {};
  for (const token of line.replace(/^EASYMVP_SMOKE\s+/, "").split(/\s+/)) {
    const separator = token.indexOf("=");
    if (separator <= 0) {
      continue;
    }
    fields[token.slice(0, separator)] = token.slice(separator + 1);
  }
  return fields;
}

function assertSmoke(condition, message, stdout, stderr) {
  if (condition) {
    return;
  }
  throw new Error(`${message}\nstdout:\n${stdout}\nstderr:\n${stderr}`);
}
