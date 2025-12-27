// Simple test runner for app.js
// Mock global document for Node.js
global.document = {
    getElementById: (id) => ({ innerText: '' })
};

import { DashboardClient } from './app.js';

function assert(condition, message) {
    if (!condition) {
        throw new Error(message || "Assertion failed");
    }
}

try {
    console.log("Running Phase 1 Tests...");
    
    const client = new DashboardClient();
    assert(client.hash === null, "Initial hash should be null");
    assert(typeof client.poll === 'function', "poll should be a function");
    assert(client.retryDelay === 1000, "Initial retry delay should be 1000ms");
    
    // Test poll function
    console.log("Testing poll()...");
    let fetchCalled = false;
    let requestHeaders = {};
    
    global.fetch = async (url, options) => {
        fetchCalled = true;
        requestHeaders = options.headers;
        return {
            ok: true,
            status: 200,
            json: async () => ({ hash: 'new-hash', updates: {} })
        };
    };

    // Trigger one poll cycle (should be async)
    // We'll mock poll to only run once for testing
    const originalPoll = client.poll;
    let pollFinished = false;
    
    // We'll test the core logic of one fetch call
    await client.fetchUpdates();
    
    assert(fetchCalled, "fetch should have been called");
    assert(requestHeaders['X-Dashboard-Hash'] === '', "Initial X-Dashboard-Hash should be empty string");
    assert(client.hash === 'new-hash', "Hash should be updated after successful fetch");

    // Test 304 Not Modified
    console.log("Testing 304 response...");
    global.fetch = async () => ({ status: 304, ok: true });
    await client.fetchUpdates();
    assert(client.hash === 'new-hash', "Hash should remain 'new-hash' after 304");

    // Test Backoff
    console.log("Testing backoff...");
    let setTimeoutCalled = 0;
    let lastDelay = 0;
    global.setTimeout = (fn, delay) => {
        setTimeoutCalled++;
        lastDelay = delay;
        fn();
    };

    global.fetch = async () => { throw new Error("Network Error"); };
    
    await client.fetchUpdates();
    assert(client.retryDelay === 2000, "Retry delay should double to 2000ms");
    assert(lastDelay === 1000, "Should have slept for 1000ms (previous delay)");

    await client.fetchUpdates();
    assert(client.retryDelay === 4000, "Retry delay should double to 4000ms");

    // Test reset backoff
    global.fetch = async () => ({ ok: true, json: async () => ({}) });
    await client.fetchUpdates();
    assert(client.retryDelay === 1000, "Retry delay should reset to 1000ms after success");

    // Test updateDOM
    console.log("Testing updateDOM()...");
    let patchSectionCalled = false;
    client.patchSection = (id, data) => {
        patchSectionCalled = true;
        assert(id === 'weather', "Should call patchSection with correct ID");
        assert(data.temp === 25, "Should call patchSection with correct data");
    };

    client.updateDOM({
        'weather': { temp: 25 }
    });
    assert(patchSectionCalled, "patchSection should have been called");

    // Test recursive patching
    console.log("Testing recursive patchSection()...");
    delete client.patchSection; // Restore original

    const fieldTemp = { dataset: { field: 'temp' }, innerText: '', classList: { add:()=>{}, remove:()=>{} } };
    const fieldCity = { dataset: { field: 'city' }, innerText: '', classList: { add:()=>{}, remove:()=>{} } };
    const fieldNested = { dataset: { field: 'nested.value' }, innerText: '', classList: { add:()=>{}, remove:()=>{} } };

    const mockElement = {
        innerText: '',
        classList: { add:()=>{}, remove:()=>{} },
        querySelectorAll: (selector) => {
            if (selector === '[data-field]') {
                return [fieldTemp, fieldCity, fieldNested];
            }
            return [];
        }
    };
    
    global.document.getElementById = () => mockElement;

    client.patchSection('weather', {
        temp: 20,
        city: 'London',
        nested: { value: 'foo' }
    });

    assert(fieldTemp.innerText === '20', "Temp should be updated to 20");
    assert(fieldCity.innerText === 'London', "City should be updated to London");
    assert(fieldNested.innerText === 'foo', "Nested value should be updated to foo");

    // Test property types and attributes
    console.log("Testing updateElement types and attributes...");
    const fieldHtml = { dataset: { field: 'content', fieldType: 'html' }, innerHTML: '', innerText: '' };
    const fieldAttr = { dataset: { field: 'icon', fieldType: 'attr', fieldAttr: 'src' }, src: '', innerText: '' };

    client.updateElement(fieldHtml, '<b>Bold</b>');
    assert(fieldHtml.innerHTML === '<b>Bold</b>', "HTML should be updated");

    client.updateElement(fieldAttr, 'sunny.png');
    assert(fieldAttr.src === 'sunny.png', "Attribute 'src' should be updated");

    // Test array to HTML
    console.log("Testing array to HTML conversion...");
    const fieldList = { dataset: { field: 'items', fieldType: 'html' }, innerHTML: '', innerText: '' };
    const items = [
        { title: 'News 1', link: 'http://foo', pub_date: 'today' }
    ];
    client.updateElement(fieldList, items);
    assert(fieldList.innerHTML.includes('News 1'), "List should contain news title");
    assert(fieldList.innerHTML.includes('http://foo'), "List should contain news link");

    // Test visual feedback
    console.log("Testing visual feedback trigger...");
    const classList = new Set();
    const flashElement = {
        innerText: '',
        dataset: { field: 'val' },
        querySelectorAll: () => [ { dataset: { field: 'val' }, innerText: 'old' } ],
        classList: {
            add: (c) => classList.add(c),
            remove: (c) => classList.delete(c)
        },
        offsetWidth: 100
    };
    
    global.document.getElementById = () => flashElement;
    
    // Simulate change
    client.patchSection('news', { val: 'new' });
    assert(classList.has('updated-flash'), "Should have added updated-flash class");

    console.log("✅ All tests passed!");
} catch (e) {
    console.error("❌ Test failed:", e.message);
    process.exit(1);
}
