import sys

if __name__ == "__main__":
    result = ""
    with open(".env", "r+") as env_file:
        parameters = ["-D" + line.rstrip() for line in env_file]
        result = " ".join(["./mvnw", *parameters, *sys.argv[1::]])
    print(result)
