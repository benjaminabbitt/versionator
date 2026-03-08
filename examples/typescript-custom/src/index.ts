import { VERSION, GIT_HASH, BUILD_DATE } from './version.js';

function main(): void {
    console.log('Sample TypeScript Application (Custom Template)');
    console.log(`Version: ${VERSION}`);
    console.log(`Git: ${GIT_HASH}`);
    console.log(`Built: ${BUILD_DATE}`);
}

main();
