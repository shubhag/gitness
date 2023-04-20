import React, { useMemo } from 'react'
import {
  Button,
  ButtonVariation,
  Color,
  Container,
  FlexExpander,
  Heading,
  Layout,
  useToggle,
  Utils
} from '@harness/uicore'
import { Else, Match, Render, Truthy } from 'react-jsx-match'
import { useHistory } from 'react-router-dom'
import { SourceCodeViewer } from 'components/SourceCodeViewer/SourceCodeViewer'
import type { OpenapiContentInfo, RepoFileContent } from 'services/code'
import {
  CodeIcon,
  decodeGitContent,
  findMarkdownInfo,
  GitCommitAction,
  GitInfoProps,
  isRefATag,
  makeDiffRefs
} from 'utils/GitUtils'
import { filenameToLanguage } from 'utils/Utils'
import { useAppContext } from 'AppContext'
import { LatestCommitForFile } from 'components/LatestCommit/LatestCommit'
import { PipeSeparator } from 'components/PipeSeparator/PipeSeparator'
import { CommitModalButton } from 'components/CommitModalButton/CommitModalButton'
import { useStrings } from 'framework/strings'
import { Readme } from '../FolderContent/Readme'
import { GitBlame } from './GitBlame'
import css from './FileContent.module.scss'

export function FileContent({
  repoMetadata,
  gitRef,
  resourcePath,
  resourceContent
}: Pick<GitInfoProps, 'repoMetadata' | 'gitRef' | 'resourcePath' | 'resourceContent'>) {
  const { routes } = useAppContext()
  const { getString } = useStrings()
  const history = useHistory()
  const [showGitBlame, toggleGitBlame] = useToggle(false)
  const content = useMemo(
    () => decodeGitContent((resourceContent?.content as RepoFileContent)?.data),
    [resourceContent?.content]
  )
  const markdownInfo = useMemo(() => findMarkdownInfo(resourceContent), [resourceContent])

  return (
    <Layout.Vertical spacing="small">
      <LatestCommitForFile repoMetadata={repoMetadata} latestCommit={resourceContent.latest_commit} standaloneStyle />
      <Container className={css.container} background={Color.WHITE}>
        <Layout.Horizontal padding="small" className={css.heading}>
          <Heading level={5} color={Color.BLACK}>
            {resourceContent.name}
          </Heading>
          <FlexExpander />
          <Layout.Horizontal spacing="xsmall">
            <Button
              variation={ButtonVariation.ICON}
              icon={CodeIcon.Edit}
              tooltip={isRefATag(gitRef) ? getString('editNotAllowed') : getString('edit')}
              tooltipProps={{ isDark: true }}
              disabled={isRefATag(gitRef)}
              onClick={() => {
                history.push(
                  routes.toCODEFileEdit({
                    repoPath: repoMetadata.path as string,
                    gitRef,
                    resourcePath
                  })
                )
              }}
            />
            <Button
              variation={ButtonVariation.ICON}
              tooltip={getString('copy')}
              icon={CodeIcon.Copy}
              tooltipProps={{ isDark: true }}
              onClick={() => Utils.copy(content)}
            />
            <CommitModalButton
              variation={ButtonVariation.ICON}
              icon={CodeIcon.Delete}
              disabled={isRefATag(gitRef)}
              tooltip={getString(isRefATag(gitRef) ? 'deleteNotAllowed' : 'delete')}
              tooltipProps={{ isDark: true }}
              repoMetadata={repoMetadata}
              gitRef={gitRef}
              resourcePath={resourcePath}
              commitAction={GitCommitAction.DELETE}
              commitTitlePlaceHolder={getString('deleteFile').replace('__path__', resourcePath)}
              onSuccess={(_commitInfo, newBranch) => {
                if (newBranch) {
                  history.replace(
                    routes.toCODECompare({
                      repoPath: repoMetadata.path as string,
                      diffRefs: makeDiffRefs(repoMetadata?.default_branch as string, newBranch)
                    })
                  )
                } else {
                  history.push(
                    routes.toCODERepository({
                      repoPath: repoMetadata.path as string,
                      gitRef
                    })
                  )
                }
              }}
            />
            <PipeSeparator />
            <Container padding={{ left: 'small', right: 'xsmall' }}>
              <Button
                variation={ButtonVariation.SECONDARY}
                text={showGitBlame ? 'View File' : 'Blame'}
                onClick={toggleGitBlame}
              />
            </Container>
          </Layout.Horizontal>
        </Layout.Horizontal>

        <Render when={(resourceContent?.content as RepoFileContent)?.data}>
          <Container className={css.content}>
            <Match expr={showGitBlame}>
              <Truthy>
                <GitBlame repoMetadata={repoMetadata} resourcePath={resourcePath} />
              </Truthy>
              <Else>
                <Render when={!markdownInfo}>
                  <SourceCodeViewer language={filenameToLanguage(resourceContent?.name)} source={content} />
                </Render>
                <Render when={markdownInfo}>
                  <Readme
                    metadata={repoMetadata}
                    readmeInfo={markdownInfo as OpenapiContentInfo}
                    contentOnly
                    maxWidth="calc(100vw - 346px)"
                    gitRef={gitRef}
                  />
                </Render>
              </Else>
            </Match>
          </Container>
        </Render>
      </Container>
    </Layout.Vertical>
  )
}