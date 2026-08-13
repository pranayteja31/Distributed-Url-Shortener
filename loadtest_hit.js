import http from 'k6/http';
import { check } from 'k6';

export const options = {
    vus: 5,
    duration: '30s',
};

export default function () {
    const response = http.get(
        'http://host.docker.internal:8082/JW9F47',
        {
            redirects: false,
        }
    );

    check(response, {
        'redirect returned': (r) =>
            r.status >= 300 && r.status < 400,
    });
}