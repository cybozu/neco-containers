// Pins packages to the versions in upstream's package-lock.json: each is
// declared on the root manifest and forced wherever it appears in the graph.
// Only names are listed here, so bumping upstream fails `--frozen-lockfile`
// until `make import-lockfile` is re-run.
const PINNED = [
  // imported by upstream but missing from its package.json
  '@protobuf-ts/runtime',
  '@protobuf-ts/runtime-rpc',
  '@types/node',
  'mobx-react-lite',
  // duplicates break instanceof (sass) or typing (d3, TS2345 in src/ui/d3)
  'sass',
  '@types/d3-selection',
  '@types/d3-zoom',
];

const DEP_FIELDS = ['dependencies', 'devDependencies', 'peerDependencies', 'optionalDependencies'];

function readPackage(pkg) {
  // not top-level: pnpm evaluates this file even where package-lock.json is absent
  const packages = require('./package-lock.json').packages;
  for (const name of PINNED) {
    const version = packages[`node_modules/${name}`]?.version;
    if (!version) throw new Error(`.pnpmfile.cjs: ${name} not found in upstream package-lock.json`);
    if (!pkg.dist) (pkg.devDependencies ??= {})[name] = version; // root manifest has no dist
    for (const field of DEP_FIELDS) {
      if (pkg[field]?.[name]) pkg[field][name] = version;
    }
  }
  return pkg;
}

module.exports = { hooks: { readPackage } };
