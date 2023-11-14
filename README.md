# simver
simple, pr based, semver git tagging logic


# definitions

## when is mmrt valid?
1. it exists
2. it is higher than the mrlt


## the most recent reserved tag (mrrt)

inside this commits tree, the highest 'vX.Y.Z-reserved' tag is the mrrt.


## the most recent live tag (mrlt)

inside this commits tree, the highest 'vX.Y.Z' tag is the mrlt.

## the my most recent tag (mmrt)

inside this commits tree, the highest 'vX.Y.Z-pr.N+base' tag is the mmrt.

## next valid tag (nvt)

find the latest mrrt or mrlt and increment the patch number (or minor if the base branch is main)

## the my most recent build number (mmrbn)

inside this commits tree, the highest '*-pr.N+(this)' is the mmrbn.

note each of the nvt, mrrt, mrlt, and mmrt are saved as valid semvers, so "X.Y.Z"
the mmrbn is an integer, so "N"

-----------

two bugs:
- need to make sure merges do not have build numbers
- need to make sure that build nums are picked up


# process

## when to run the workflow
1. for every commit on main
2. for every pr event

### 1. for each pr event:
1. figure out if head commit is a pr merge, commit, or nothing
 - check if the current head has a semver tag
    - if it does, then it is a nothing
 - check if the commit message contains "(#N)" where N is a number and the pr exists and the head commit is the current commit
 - if it is, then it is a pr merge
 - if it is not, then it is a normal commit

#### 1. if nothing:
1. do nothing

#### 2. if pr merge:
1. find the mrrt, mrlt, and mmrt
2. check if the mmrt is valid
3. if it is, move forward with the mmrt
4. if it is not, use the nvt
 - create two new tags on the base commit
    - vX.Y.Z-reserved
    - vX.Y.Z-pr.N+base
5. create a new tag on the head commit

#### 3. if a normal commit:
1. find the mrrt and mrlt, calculate the nvt
2. create a new tag (based on nvt) on the head commit

### 2. for each commit on main:
1. figure out if head is a pr merge or not
 - check if the commit message contains "(#N)" where N is a number and the pr exists and the head commit is the current commit
 - if it is, then it is a pr merge
 - if it is not, then it is a normal commit

#### 1. if pr merge:
1. find the mrrt, mrlt, and mmrt
2. check if the mmrt is valid
3. if it is, move forward with the mmrt
4. if it is not, use the nvt
 - create two new tags on the base commit (for brevity) - get the base commit from the pr
    - vX.Y.Z-reserved
    - vX.Y.Z-pr.N+base
5. create a new tag on the head commit with no build number or prerelease

#### 2. if a normal commit:
1. find the mrrt and mrlt, calculate the nvt
2. create a new tag (based on nvt) on the head commit with no build number or prerelease

