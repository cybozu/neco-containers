// Pins packages to the versions in upstream's package-lock.json. Only names
// are listed; after bumping upstream, re-run `make import-lockfile`.

// Imported by upstream but not declared in its package.json.
const PHANTOM = [
  '@protobuf-ts/runtime',
  '@protobuf-ts/runtime-rpc',
  '@types/node',
  'mobx-react-lite',
];
// Without the pin these resolve to a duplicate (sass) or newer (d3 types) version.
const OVERRIDE = [
  'sass',
  '@types/d3-selection',
  '@types/d3-zoom',
];
const PINNED = [...PHANTOM, ...OVERRIDE];

const DEP_FIELDS = ['dependencies', 'devDependencies', 'peerDependencies', 'optionalDependencies'];

function readPackage(pkg) {
  // required here, not at top level: pnpm loads this file for every command,
  // also where package-lock.json is absent (e.g. `pnpm config get` in hubble-ui/)
  const packages = require('./package-lock.json').packages;
  const upstream = packages['']; // upstream's package.json as recorded in its lock
  const isRoot = pkg.name === upstream.name && pkg.private === true;
  for (const name of PINNED) {
    const version = packages[`node_modules/${name}`]?.version;
    if (!version) throw new Error(`.pnpmfile.cjs: ${name} not found in upstream package-lock.json`);
    if (PHANTOM.includes(name) && (upstream.dependencies?.[name] || upstream.devDependencies?.[name])) {
      throw new Error(`.pnpmfile.cjs: ${name} is now declared in upstream package.json; remove it from PHANTOM`);
    }
    if (isRoot) (pkg.devDependencies ??= {})[name] = version;
    for (const field of DEP_FIELDS) {
      if (pkg[field]?.[name]) pkg[field][name] = version;
    }
  }
  return pkg;
}

module.exports = { hooks: { readPackage } };
