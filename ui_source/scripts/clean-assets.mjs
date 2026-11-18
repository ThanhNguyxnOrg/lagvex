// Removes stale hashed Vite entry bundles from web/assets.
//
// The build writes content-hashed names (index-<hash>.js / .css) and the
// output directory is kept non-empty on purpose (static game art lives
// there), so old bundles would otherwise pile up and stay embedded in the
// Go binary. Only the generated entry files are touched; every other asset
// in the directory is left alone.
import { readdir, unlink } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const here = path.dirname(fileURLToPath(import.meta.url));
const assetsDir = path.resolve(here, "..", "..", "web", "assets");
const entryPattern = /^index-[A-Za-z0-9_-]+\.(js|css)$/;

const removed = [];
for (const name of await readdir(assetsDir)) {
  if (!entryPattern.test(name)) continue;
  await unlink(path.join(assetsDir, name));
  removed.push(name);
}

if (removed.length > 0) {
  console.log(`cleaned ${removed.length} stale entry bundle(s): ${removed.join(", ")}`);
}
