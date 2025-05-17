#!/usr/bin/env python3
#
#

from optparse import OptionParser
from datetime import datetime

import logging
import re
import sys
import json
import time
import urllib.request
import logging.handlers

logging.basicConfig(level=logging.INFO,
                    format='[%(asctime)s] %(levelname)-8s (%(name)s) %(message)s')
logger = logging.getLogger(__name__)

TIMEOUT = 5


def request_releases(repo):
    '''
        Pull a paginated list of release of this repo 
    '''
    current_url = f"https://api.github.com/repos/{repo}/releases?per_page=100"

    logger.info(f'polling {current_url} for latest release')

    versions = []

    while True:

        header = {'Accept': 'application/vnd.github.v3+json'}

        req = urllib.request.Request(
            url=current_url, headers=header, method='GET')
        res = urllib.request.urlopen(req, timeout=TIMEOUT)

        res_body = res.read()

        current_versions = json.loads(res_body.decode("utf-8"))

        # append our results
        versions.extend(current_versions)

        # do we have another page
        if res.headers.get('link', None):

            current_url = extract_link(res.headers.get('link'))

            if current_url is None:
                # end of pages
                break

            logger.debug(f'polling {current_url} for next page')

            # rate limit un-authed queries
            time.sleep(0.1)
        else:
            break

    return versions


def extract_link(link_header):
    '''
        Used to extract the next page url from the link: header
    '''

    logger.debug('extract_link')

    for link in link_header.split(','):
        # If there is a 'next' link return the URL between the angle brackets, or None
        if 'rel="next"' in link:
            return link[link.find("<")+1:link.find(">")]

    return None


def sort_releases(releases):
    '''
        return a date ordered list of releases
    '''

    logger.debug('sort_releases')

    return sorted(
        releases,
        key=lambda release: datetime.fromisoformat(
            release['created_at'].replace("Z", "+00:00")),
        reverse=True
    )


def filter_versions(versions, filter_regex):
    '''
        Filter a list of release with a regex , optionally filter based
        on a regex capture
    '''

    logger.debug('filter_versions')

    p = re.compile(filter_regex)

    filtered_versions = []

    for version in versions:

        match = p.match(version)

        if match:
            # are we using a capture group
            if len(match.groups()):
                filtered_versions.append(match.group(1))
            else:
                filtered_versions.append(version)

    return filtered_versions


def poll_versions(repo, filter_regex, all_versions):

    logger.debug('poll_version')

    stable_version = None

    version_data = request_releases(repo)

    # filter pre-releases
    full_version_data = list(
        filter(lambda x: x['prerelease'] == False, version_data))

    # sort by release date
    full_version_data = sort_releases(full_version_data)

    # extract version from version data (use tag_name in preference to name)
    versions = list(map(lambda x: x['tag_name'], full_version_data))

    # apply our regex
    candidate_versions = filter_versions(versions, filter_regex)

    if all_versions:
        return candidate_versions

    return [candidate_versions.pop(0)]


def poll(options):

    logger.debug('poll')

    logger.info('polling {} for latest release'.format(options.repo))

    versions = poll_versions(
        options.repo, options.filter_regex, options.all_versions)

    if len(versions):

        logger.info('updating version file {} with: version(s): {}'.format(
            options.version_file, versions))

        with open(options.version_file, 'w') as version_file:
            for version in versions:
                version_file.write(f"{version}\n")

    return


def main():

    parser = OptionParser()

    parser.add_option("-r", "--repo", dest="repo",
                      help="repo to poll", metavar="REPO")

    parser.add_option("--filterregex", dest="filter_regex", default=".*",
                      help="regex to filter versions", metavar="REGEX")

    parser.add_option("-f", "--file", dest="version_file", default="release.txt",
                      help="version file to write out to", metavar="FILE")

    parser.add_option("-a", "--all", action="store_true", dest="all_versions", default=False,
                      help="return all current released versions rather than just the latest", metavar="ALL")

    parser.add_option("-v", "--verbose",
                      action="store_true", dest="verbose", default=False,
                      help="print status messages to stdout")

    (options, args) = parser.parse_args()

    if options.repo is None:
        parser.error("github repo not supplied")

    if options.verbose:
        logger.setLevel(logging.DEBUG)

    return poll(options)


if __name__ == "__main__":
    sys.exit(main())
