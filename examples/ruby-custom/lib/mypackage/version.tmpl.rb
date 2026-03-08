# frozen_string_literal: true
# Custom version file for mypackage

module Mypackage
  VERSION = "{{MajorMinorPatch}}"
  FULL_VERSION = "{{MajorMinorPatch}}{{PreReleaseWithDash}}{{MetadataWithPlus}}"
  GIT_HASH = "{{ShortHash}}"
  BUILD_DATE = "{{BuildDateTimeUTC}}"
end
