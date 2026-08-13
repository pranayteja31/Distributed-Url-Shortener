import http from 'k6/http';
import { check } from 'k6';

const urls = [
    'RYKPer',
    'ZDyHWF',
    'TA8tgJ',
    'JW9F47',
    'cVVUSZ',
];

export const options = {
    vus: 5,
    duration: '30s',
};

export default function () {
    const shortCode = urls[Math.floor(Math.random() * urls.length)];

    const response = http.get(
        `http://host.docker.internal:8082/${shortCode}`,
        {
            redirects: false,
        }
    );

    check(response, {
        'redirect returned': (r) =>
            r.status >= 300 && r.status < 400,
    });
}