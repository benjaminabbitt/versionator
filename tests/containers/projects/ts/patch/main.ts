// Read version from package.json
import pkg from './package.json' with { type: 'json' };
console.log('Version:', pkg.version);
