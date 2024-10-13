V2
--

After using goasync for some time, I think there are some possible features that I want for
a v2 which will cause breaking changes. These currently include:

    * Allow for specific tasks to finish and cause the rest of the tasks to stop without returning any errors
        * This feature itself might not requre v2 changes. Could have a `AddStopExecuteTags(name string, task ExecuteTask) error`.
          With this only being able to be called multiple times? Then when all process have stopped it is done. Also, can only
          be called before Run() is invoked. Need to think more about the possible use cases
    * Pass the context to both the Initalize() and Cleanup() functions