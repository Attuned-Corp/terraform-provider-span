terraform {
  required_providers {
    span = {
      source = "registry.terraform.io/attuned-corp/span"
    }
  }
}

provider "span" {
  access_token = "<your PAT>"
}

#======================
# span_person allows for querying a person resource from within span
#
# data "span_person" "john_smith" {
#   email = "john@smith.com"
# }

# ## Example resource:
# data "span_person" "john_smith" {
#     email = "john@smith.com"
#     name  = "John Smith"
#     teams = [
#         {
#             id   = "ccbed53f-0e3c-488b-878e-2c4cfb131e5d"
#             name = "Team 1"
#         },
#         {
#             id   = "049f1f94-f638-4284-b435-b2e998980b81"
#             name = "Team 2"
#         },
#     ]
# }
#======================


#======================
# span_people loads all people with additional optional filtering.
#
# data "span_people" "all" {
#   team_ids = ["<id-1>", "<id-2>"]
# }
#
# ## Example resource:
# data "span_people" "all" {
#    people = [
#        {
#            email = "john@smith.com"
#            name  = "John Smith",
#            teams = [
#                {
#                    id   = "ccbed53f-0e3c-488b-878e-2c4cfb131e5d"
#                    name = "Team 1"
#                },
#                {
#                    id   = "049f1f94-f638-4284-b435-b2e998980b81
#                    name = "Team 2"
#                },
#            ]
#        },
#        {
#            email = "jane@doe.com"
#            name  = "Jane Doe"
#            teams = [
#                {
#                    id   = "ccbed53f-0e3c-488b-878e-2c4cfb131e5d"
#                    name = "Team 1"
#                },
#            ]
#        },
#======================

#======================
# span_team provides information on an individual team by id or name
#
# data "span_team" "platform" {
#   name = "Platform"
#   # or
#   id = "<team id>"
# }
#
# ## Example resource:
#
# data "span_team" "platform" {
#   id      = "5bbed53f-0e3c-488b-878e-2c4cfb131e5d"
#   name    = "Team 1"
#   slug    = "team-1"
#    members = [
#        {
#            email     = "john@smith.com"
#            name      = "John Smith"
#            team_lead = true
#        },
#        {
#            email     = "jane@doe.com"
#            name      = "Jane Doe"
#            team_lead = false
#        },
#    ]
#}
#
#======================


#======================
# span_teams loads all team resources with minimal view of each
#
# data "span_teams" "all" {}
#
# ## Example resource:
# data "span_teams" "all" {
#     teams = [
#         {
#             id   = "ccbed53f-0e3c-488b-878e-2c4cfb131e5d"
#             name = "Team 1"
#             slug = "team-1"
#         },
#         {
#             id   = "049f1f94-f638-4284-b435-b2e998980b81"
#             name = "Team 2"
#             slug = "team-2"
#         },
#     ]
# }
