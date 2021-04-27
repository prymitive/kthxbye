#!/usr/bin/env python3


import json
import os
import sys
from datetime import datetime
from urllib.request import Request, urlopen


def apiRequest(token, path):
    api = "api.github.com"
    uri = "https://{api}{path}".format(api=api, path=path)

    req = Request(uri)
    req.add_header("Authorization", "token %s" % token)
    req.add_header("Content-Type", "application/json")

    return req


def shouldDelete(package):
    created_at = package["created_at"]
    if created_at[-1] == "Z":
        created_at = created_at[:-1] + "+00:00"
    created = datetime.fromisoformat(created_at).replace(tzinfo=None)
    delta = datetime.now() - created
    if (
        package["metadata"]["package_type"] == "container"
        and len(package["metadata"]["container"]["tags"]) == 0
        and delta.total_seconds() > 3600 * 24
    ):
        return True
    return False


def deletePackage(package):
    print(
        "Will purge package {} created at {}".format(
            package["id"], package["created_at"]
        )
    )
    req = apiRequest(
        token, "/user/packages/container/kthxbye/versions/{}".format(package["id"])
    )
    req.method = "DELETE"
    try:
        response = urlopen(req)
    except Exception as e:
        print("DELETE request to '%s' failed: %s" % (req.get_full_url(), e))


def purge(token):
    req = apiRequest(token, "/user/packages/container/kthxbye/versions?per_page=100")
    try:
        response = urlopen(req)
    except Exception as e:
        print("GET request to '%s' failed: %s" % (req.get_full_url(), e))
    else:
        packages = json.load(response)
        for package in packages:
            if shouldDelete(package):
                deletePackage(package)


if __name__ == "__main__":
    token = os.getenv("GITHUB_TOKEN")

    if not token:
        print("GITHUB_TOKEN env variable is missing")
        sys.exit(1)

    purge(token)
