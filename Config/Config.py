# -*- coding: utf-8 -*-
'''
Created on 16 sept. 2023

@author: BoxBoxJason
'''
import os
import logging
import sys

def getConfig(configDict):
    wordsToSearchFor = configDict.get("wordsToSearchFor",[])
    resultDirPath = os.path.realpath(configDict.get("resultDir",os.path.join(__file__,"..","..","Results")))
    outputTextRanking = configDict.get("outputTextRanking",True)
    if os.path.exists(resultDirPath) and os.path.isdir(resultDirPath) and os.access(resultDirPath,os.W_OK):
        os.environ['RESULTSDIRPATH'] = resultDirPath
    else:
        logging.fatal(f"resultDirPath must be a directory with Write permissions, check {resultDirPath}")
        sys.exit(1)

    return wordsToSearchFor,resultDirPath,outputTextRanking
