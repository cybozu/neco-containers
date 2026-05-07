// readPackage hook injects phantom devDeps that upstream's webpack imports
// but doesn't declare, and overrides version specifiers to dedupe packages
// whose duplicate copies break instanceof (e.g. sass.SassString across two
// sass copies pulled by @blueprintjs/node-build-scripts and sass-loader).
//
// The hook fires for both registry-fetched packages and the local root
// manifest; we discriminate by `pkg.dist`, which exists only on registry
// entries (root manifests have no dist).
const PHANTOM_DEV_DEPS = {
  '@protobuf-ts/runtime': '2.11.1',
  '@protobuf-ts/runtime-rpc': '2.11.1',
  '@types/node': '24.1.0',
  'mobx-react-lite': '4.1.0',
};
const OVERRIDE = {
  sass: '1.89.2',
};

const DEP_FIELDS = ['dependencies', 'devDependencies', 'peerDependencies', 'optionalDependencies'];

function readPackage(pkg) {
  if (!pkg.dist) {
    pkg.devDependencies = { ...pkg.devDependencies, ...PHANTOM_DEV_DEPS };
  }
  for (const field of DEP_FIELDS) {
    if (!pkg[field]) continue;
    for (const [name, version] of Object.entries(OVERRIDE)) {
      if (pkg[field][name]) pkg[field][name] = version;
    }
  }
  return pkg;
}

module.exports = { hooks: { readPackage } };
