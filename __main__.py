# -*- coding: utf-8 -*-
'''
Created on 3 mars 2023

Creates Instagram messages statistics from your conversations
Input conversation data collected from instagram and builds a complete set of graphs and .txt files
Before launching, you must download your data from Instagram (.json format)

Can be used to analyse one to one conversations, group conversations and even mix several different conversations

LAUNCH:
    Takes one or two arguments:
        - First argument (mandatory) is the path to messages source folder (the messages must be directly below the folder)
        - Second argument (optional) is the path to the config.json file, by default it's in Config/config.json


CONFIG:
    In the Config/config.json, you can select 3 options:
        - resultDirPath [string] (optional): the path to a folder were results will be outputed, default: Results
        - wordsToSearchFor [list] (optional) : list of words, default: []
        - outputTextRanking [bool] (optional) : defines if output should include
                            text rankings (easier to read for large group conversations), default: True

GRAPHS:

    Occurences of vocabulary (optionnal):
        - Pie chart comparing occurences of chosen word(s)
        Shows how many times a word has been used by users
        Can be abled / disabled in config.json file: words to search for are inputed as a list
        - .txt ranking (better for large groups)
    
    Messages count:
        - Pie chart comparing number of messages by user
        - .txt ranking (better for large groups)
    
    Reactions count:
        - Pie chart comparing number of reactions by user
        - .txt ranking (better for large groups)

    Reels count:
        - Pie chart comparing number of reels sent by user
        - .txt ranking (better for large groups)

    Photos / Videos count:
        - Pie chart comparing number of photos and videos sent by user
        - .txt ranking (better for large groups)
'''
import json
import logging
import os
import sys
import time

from Process.CollectData import collectData
from Process.ProcessStats import processData,processTimeStats
from Config.Config import getConfig


##----------Logging setup----------##
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[
        logging.FileHandler(os.path.join(os.path.dirname(__file__),"logging.log")),
        logging.StreamHandler(sys.stdout)
    ]
)

if len(sys.argv) < 2:
    logging.fatal("Please provide at least one argument: Path to messages source folder")
    sys.exit(1)
elif len(sys.argv) > 3:
    logging.fatal("Please provide at most two arguments: Path to messages source folder AND Path to config .json file")
    sys.exit(1)
else:
    sourceFolderPath = sys.argv[1]
    configDictPath = os.path.join(os.path.dirname(__file__),"Config","config.json")
    configGiven = True

    if not (os.path.exists(sourceFolderPath) and os.path.isdir(sourceFolderPath) and os.access(sourceFolderPath,os.R_OK)):
        logging.fatal(f"Could not look into {os.path.realpath(sourceFolderPath)}, it must be a folder with read access")
        sys.exit(1)

    if len(sys.argv) == 3:
        configDictPath = sys.argv[2]

    if not (configDictPath.endswith(".json") and os.path.exists(configDictPath) \
            and os.access(configDictPath,os.R_OK)):
        logging.info(f"Could not look into {os.path.realpath(configDictPath)}, proceeding with default config")
        configGiven = False

#Creating variables for extraction
if configGiven:
    with open(configDictPath,'r',encoding='utf-8') as configFile:
        configDict = json.load(configFile)
else:
    configDict = {}

wordsToSearchFor,resultDirPath,outputTextRanking = getConfig(configDict)

from Graphs.Graphs import pieChartFigure,heatMapFigure,barGraphFigure,textRanking
beg = time.time()

logging.info("Beginning data collection")
usersDict,daysCount,gridHourDay = collectData(sourceFolderPath,wordsToSearchFor)
logging.info("Data collection successful")

logging.info("Beginning data sorting")
messages,reactions,reels,photosAndVideos,wordsToFind,vocabulary,COLORS = processData(usersDict,wordsToSearchFor)
daysList,countList,monthsList,monthsCount = processTimeStats(gridHourDay,daysCount)
logging.info("Data sorting successful")

logging.info("Generating graphs")
#Words found from list of words pie chart
if wordsToSearchFor:
    pieChartFigure(wordsToFind.count, wordsToFind.usersList,
                   "Number of words found", "Number of words found", COLORS)

#Creating vocabulary pie chart
pieChartFigure(vocabulary.count, vocabulary.usersList,
               "Vocabulary", "Vocabulary richness (words)", COLORS,len(vocabulary.totalVocab))

#Creating total messages pie chart
pieChartFigure(messages.count,messages.usersList,"Number of Messages","Number of Messages",COLORS)

#Creating total reactions pie chart
pieChartFigure(reactions.count,reactions.usersList,"Number of Reactions","Reactions to messages",COLORS)

#Creating reels sent pie chart
pieChartFigure(reels.count,reels.usersList,"Number of Reels sent","Reels sent",COLORS)

#Creating photos and videos sent pie chart
pieChartFigure(photosAndVideos.count,photosAndVideos.usersList,
               "Number of photos and videos sent","Photos and videos sent",COLORS)


## TEMPORAL STATISTICS ##

heatMapFigure(gridHourDay, "Activity by time and day", "Activity by time and day", "Day", "Time")
barGraphFigure(daysList,countList,"Number of messages per day","Messages per day","Date","Number of messages")
barGraphFigure(monthsList, monthsCount, "Number of messages per month",
                "Messages per month", "Month", "Number of messages","#000359")

if outputTextRanking:
    #Creating total messages ranking and vocabulary ranking
    textRanking(messages.textRanking,sum(messages.count),"messages",
                "Number of messages ranking (Total {}):","Number of messages")

    textRanking(vocabulary.textRanking,len(vocabulary.totalVocab),"different words",
                "Vocabulary richness ranking (Total {}):","Vocabulary richness")

    textRanking(reactions.textRanking,sum(reactions.count),"reactions",
                "Number of reactions ranking (Total {})","Number of reactions")

    textRanking(reels.textRanking,sum(reels.count),"reels",
                "Number of reels sent ranking (Total {})","Number of reels")

    textRanking(photosAndVideos.textRanking,sum(photosAndVideos.count),"photos and videos",
                "Number of photos and videos ranking (Total {})","Number of photos and videos")

    if wordsToSearchFor:
        textRanking(wordsToFind.textRanking,sum(wordsToFind.count),"words found",
                    "Number of words found (Total {})","Number of words found")

logging.info(f"End of generation, check results in {resultDirPath}")
logging.info("Runtime {time.time()-beg}")
#pyplot.show() #uncomment to open tabs
