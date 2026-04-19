import { access } from "node:fs/promises";
import path from "node:path";

const desktopRoot = path.resolve(import.meta.dirname, "..");
const releaseRoot = path.join(desktopRoot, "release", "linux-unpacked");
const binaryPath = path.join(releaseRoot, "resources", "bin", "easymvp-core");
const executablePath = path.join(releaseRoot, "easymvp-desktop");

await access(releaseRoot);
await access(binaryPath);
await access(executablePath);

console.log(`verified desktop package at ${releaseRoot}`);
console.log(`verified bundled core binary at ${binaryPath}`);
