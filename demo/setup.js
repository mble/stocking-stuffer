const fs = require("fs");
const path = require("path");
const readline = require("readline");
const crypto = require("crypto");
const spawn = require("child_process").spawn;
const bcrypt = require("bcrypt");
let reader = readline.createInterface({
  input: fs.createReadStream(path.join(__dirname, "combo.txt"), {
    encoding: "utf-8"
  }),
  console: false
});

reader.on("line", line => {
  let res = line.split(":");
  let username = res[0];
  let password = res[1];
  bcrypt.hash(password, 8, (err, hash) => {
    spawn("psql", [
      "-d",
      "stocking-stuffer_dev",
      "-c",
      `insert into users (username, password) values ('${username}', '${hash}')`
    ]);
  });

  if (Math.floor(Math.random() * (10 - 1) + 1) >= 7) {
    let pass_hash = crypto
      .createHash("md5")
      .update(password)
      .digest("hex");
    spawn("psql", [
      "-d",
      "stocking-stuffer_dev",
      "-c",
      `insert into pwned_passwords (password_md5_hash) values ('${pass_hash}')`
    ]);
  }
});
