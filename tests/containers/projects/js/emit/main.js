#!/usr/bin/env node
// Test application that imports generated version module

import { VERSION, MAJOR, MINOR, PATCH } from './version.js';

console.log(`Version: ${VERSION}`);
console.log(`Major: ${MAJOR}, Minor: ${MINOR}, Patch: ${PATCH}`);
