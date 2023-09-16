# -*- coding: utf-8 -*-
'''
Created on 15 sept. 2023

@author: BoxBoxJason
'''
from datetime import timedelta
import logging
from random import randint
from Resources import enums


def processData(usersDict,wordsToSearchFor):
    # List of colors by user index
    COLORS = [f"#{randint(0, 0xFFFFFF):06x}" for i in range(len(usersDict))]

    messages = StatisticsGrouper()
    reactions = StatisticsGrouper()
    reels = StatisticsGrouper()
    photosAndVideos = StatisticsGrouper()
    wordsToFind = StatisticsGrouper()
    vocabulary = StatisticsGrouper()

    logging.debug("Sorting users data")
    for userDict in usersDict.values():
        buildCount(messages, userDict, 'messages', 'messages')
        buildCount(reactions,userDict,'reactions','reactions')
        buildCount(reels,userDict,'reels','reels')
        buildCount(photosAndVideos,userDict,'photosAndVideos','photos and videos')
        buildCount(wordsToFind,userDict,'wordsToFind','words')
        buildCount(vocabulary,userDict,'vocabulary','different words')

        vocabulary.totalVocab = vocabulary.totalVocab | userDict["vocabulary"]

    messages.textRanking = buildRanking(usersDict, 'messages')
    reactions.textRanking = buildRanking(usersDict,'reactions')
    reels.textRanking = buildRanking(usersDict,'reels')
    photosAndVideos.textRanking = buildRanking(usersDict,'photosAndVideos')
    wordsToFind.textRanking = buildRanking(usersDict,'wordsToFind')
    vocabulary.textRanking = buildRanking(usersDict, 'vocabulary')

    return messages,reactions,reels,photosAndVideos,wordsToFind,vocabulary,COLORS


class StatisticsGrouper:

    def __init__(self):
        self.count = []
        self.usersList = []
        self.textRanking = None
        self.totalVocab = set()


def buildCount(statsGrouper,userDict,dictKey,msgtype):
    """
    Updates the count list and user list text
    """
    count = userDict[dictKey]
    if not isinstance(count,int):
        count = len(count)
    statsGrouper.count.append(count)
    statsGrouper.usersList.append(f"{userDict['name']} ({count} {msgtype})")


def buildRanking(usersDict,dictKey):
    """
    Builds a dictionary used to construct .txt ranking
    """
    countRanking = {}
    for user,userDict in usersDict.items():
        countUser = userDict[dictKey]
        if not isinstance(countUser,int):
            countUser = len(countUser)

        if not countUser in countRanking:
            countRanking[countUser] = [user]
        else:
            countRanking[countUser].append(user)

    return countRanking


def processTimeStats(gridHourDay,daysCount):
    """
    Processes time statistics of provided data
    """
    # Total number of days talked
    TOTALNUMBEROFDAYS = len(daysCount)
    # First day of conversation
    STARTDATE = min(daysCount)
    # Last day of conversation
    ENDDATE = max(daysCount)
    # List of days between start of conversation and last message of conversation
    ALLDAYS = [STARTDATE+timedelta(days=x) for x in range((ENDDATE-STARTDATE).days)]

    #Determining average message per hour / day of the week
    for i in range(len(gridHourDay)):
        for j in range(len(gridHourDay[0])):
            gridHourDay[i][j] /= TOTALNUMBEROFDAYS

    #Creating every day activity bar graph
    logging.debug("Creating messages per day bar graph")

    daysList = []
    countList = []
    for day in ALLDAYS:
        daysList.append(day)
        countList.append(daysCount.get(day,0))

    #Creating every month activity bar graph
    logging.debug("Creating messages per month bar graph")
    #Reprocessing every day activity into every month activity
    monthsList = []
    monthsCount = []
    for i,day in enumerate(daysList):
        monthYearDictKey = f"{enums.MONTHS[day.month-1]}-{day.year}"
        try:
            position = monthsList.index(monthYearDictKey)
            monthsCount[position] += countList[i]
        except ValueError:
            monthsList.append(monthYearDictKey)
            monthsCount.append(countList[i])

    return daysList,countList,monthsList,monthsCount
