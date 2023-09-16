# Instagram Messages Statistics
Creates Instagram messages statistics from your conversations.<br/>
Input conversation data collected from instagram and builds a complete set of graphs and .txt files to display statistics.<br/>
*Before launching, you must download your data from Instagram (.json format)*

Can be used to analyse one to one conversations, group conversations and even mix several different conversations

### LAUNCH:
    Takes one or two arguments:
        - First argument (mandatory) is the path to messages source folder (the messages must be directly below the folder)
        - Second argument (optional) is the path to the config.json file, by default it's in Config/config.json


### CONFIG:
    In the Config/config.json, you can select 3 options:
        - resultDirPath [string] (optional): the path to a folder were results will be outputed, default: Results
        - wordsToSearchFor [list] (optional) : list of words, default: []
        - outputTextRanking [bool] (optional) : defines if output should include
                            text rankings (easier to read for large group conversations), default: True

### GRAPHS:

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

    Activity per day:
        - Bar graph representing the number of messages sent per day

    Activity per month:
        - Bar graph representing the number of messages sent per month

    Activity per hour and day of the week:
        - Heatmap representing the time periods where most messages are sent throughout an average week
