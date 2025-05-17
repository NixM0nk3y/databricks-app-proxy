const config = {
    versions: {
        GO: "1.24.3",
        CADDY: "2.10.0",
    },
    task: {
        memory: 512,
        cpu: 256,
        count: 2,
    },
};

export default {
    ...config,
};
