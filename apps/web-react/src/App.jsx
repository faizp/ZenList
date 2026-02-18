import { useEffect, useMemo, useState } from 'react';

const GRAPHQL_URL = import.meta.env.VITE_GRAPHQL_URL || '/query';

const STATUS_OPTIONS = ['TODO', 'IN_PROGRESS', 'BLOCKED', 'DONE'];
const PRIORITY_OPTIONS = ['P1', 'P2', 'P3', 'P4', 'P5'];

const Q_ME = `
  query Me {
    me {
      id
      name
      email
      timezone
      avatarUrl
      createdAt
      updatedAt
    }
  }
`;

const Q_PROJECTS = `
  query Projects($first: Int) {
    projects(first: $first) {
      edges {
        node {
          id
          title
          description
          color
          createdAt
          updatedAt
        }
      }
    }
  }
`;

const Q_LABELS = `
  query Labels($first: Int) {
    labels(first: $first) {
      edges {
        node {
          id
          name
          createdAt
          updatedAt
        }
      }
    }
  }
`;

const Q_TASKS = `
  query Tasks($projectId: ID!, $statuses: [TaskStatus!], $priorities: [TaskPriority!], $first: Int) {
    tasks(projectId: $projectId, statuses: $statuses, priorities: $priorities, first: $first) {
      edges {
        node {
          id
          parentTaskId
          title
          description
          status
          priority
          startAt
          dueAt
          completedAt
          createdAt
          updatedAt
          labels {
            id
            name
          }
          subtasks {
            id
            parentTaskId
            title
            description
            status
            priority
            startAt
            dueAt
            completedAt
            createdAt
            updatedAt
            labels {
              id
              name
            }
          }
        }
      }
    }
  }
`;

const M_UPSERT_ME = `
  mutation UpsertMe($input: UpsertMeInput!) {
    upsertMe(input: $input) {
      id
      name
      email
      timezone
      avatarUrl
      updatedAt
    }
  }
`;

const M_CREATE_PROJECT = `
  mutation CreateProject($input: CreateProjectInput!) {
    createProject(input: $input) {
      id
      title
      description
      color
      updatedAt
    }
  }
`;

const M_UPDATE_PROJECT = `
  mutation UpdateProject($input: UpdateProjectInput!) {
    updateProject(input: $input) {
      id
      title
      description
      color
      updatedAt
    }
  }
`;

const M_DELETE_PROJECT = `
  mutation DeleteProject($id: ID!) {
    deleteProject(id: $id) {
      id
      deletedAt
    }
  }
`;

const M_CREATE_LABEL = `
  mutation CreateLabel($input: CreateLabelInput!) {
    createLabel(input: $input) {
      id
      name
      updatedAt
    }
  }
`;

const M_UPDATE_LABEL = `
  mutation UpdateLabel($input: UpdateLabelInput!) {
    updateLabel(input: $input) {
      id
      name
      updatedAt
    }
  }
`;

const M_DELETE_LABEL = `
  mutation DeleteLabel($id: ID!) {
    deleteLabel(id: $id) {
      id
      deletedAt
    }
  }
`;

const M_CREATE_TASK = `
  mutation CreateTask($input: CreateTaskInput!) {
    createTask(input: $input) {
      id
      parentTaskId
      title
      status
      priority
      startAt
      dueAt
      completedAt
    }
  }
`;

const M_UPDATE_TASK = `
  mutation UpdateTask($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      id
      parentTaskId
      title
      description
      status
      priority
      startAt
      dueAt
      completedAt
      updatedAt
    }
  }
`;

const M_DELETE_TASK = `
  mutation DeleteTask($id: ID!) {
    deleteTask(id: $id) {
      id
      deletedAt
    }
  }
`;

async function gqlRequest(query, variables = {}) {
  const response = await fetch(GRAPHQL_URL, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ query, variables })
  });

  const payload = await response.json();
  if (!response.ok || payload.errors?.length) {
    const message = payload.errors?.[0]?.message || `request failed: ${response.status}`;
    throw new Error(message);
  }
  return payload.data;
}

function toDatetimeLocal(isoString) {
  if (!isoString) return '';
  const dt = new Date(isoString);
  const pad = (value) => String(value).padStart(2, '0');
  const year = dt.getFullYear();
  const month = pad(dt.getMonth() + 1);
  const day = pad(dt.getDate());
  const hour = pad(dt.getHours());
  const min = pad(dt.getMinutes());
  return `${year}-${month}-${day}T${hour}:${min}`;
}

function toISO(localValue) {
  if (!localValue) return null;
  return new Date(localValue).toISOString();
}

function formatDate(isoString) {
  if (!isoString) return '—';
  return new Date(isoString).toLocaleString();
}

export default function App() {
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  const [me, setMe] = useState(null);
  const [projects, setProjects] = useState([]);
  const [labels, setLabels] = useState([]);
  const [tasks, setTasks] = useState([]);

  const [selectedProjectId, setSelectedProjectId] = useState('');
  const [selectedLabelId, setSelectedLabelId] = useState('');
  const [selectedTaskId, setSelectedTaskId] = useState('');

  const [meForm, setMeForm] = useState({
    name: 'ZenList User',
    email: 'user@zenlist.local',
    timezone: 'UTC',
    avatarUrl: ''
  });

  const [projectCreateForm, setProjectCreateForm] = useState({
    title: '',
    description: '',
    color: ''
  });

  const [projectUpdateForm, setProjectUpdateForm] = useState({
    title: '',
    description: '',
    color: ''
  });

  const [labelCreateName, setLabelCreateName] = useState('');
  const [labelUpdateName, setLabelUpdateName] = useState('');

  const [taskFilters, setTaskFilters] = useState({
    statuses: [],
    priorities: []
  });

  const [taskCreateForm, setTaskCreateForm] = useState({
    title: '',
    description: '',
    status: 'TODO',
    priority: 'P3',
    startAt: '',
    dueAt: '',
    labelIds: []
  });

  const [subtaskCreateForm, setSubtaskCreateForm] = useState({
    parentTaskId: '',
    title: '',
    description: '',
    status: 'TODO',
    priority: 'P3',
    startAt: '',
    dueAt: '',
    labelIds: []
  });

  const [taskUpdateForm, setTaskUpdateForm] = useState({
    title: '',
    description: '',
    status: '',
    priority: '',
    startAt: '',
    dueAt: '',
    labelIds: []
  });

  const selectedProject = useMemo(
    () => projects.find((project) => project.id === selectedProjectId) || null,
    [projects, selectedProjectId]
  );

  const rootTasks = useMemo(() => tasks || [], [tasks]);
  const allTasks = useMemo(() => {
    const list = [];
    for (const task of rootTasks) {
      list.push(task);
      for (const subtask of task.subtasks || []) {
        list.push(subtask);
      }
    }
    return list;
  }, [rootTasks]);

  const selectedTask = useMemo(
    () => allTasks.find((task) => task.id === selectedTaskId) || null,
    [allTasks, selectedTaskId]
  );

  const selectedLabel = useMemo(
    () => labels.find((label) => label.id === selectedLabelId) || null,
    [labels, selectedLabelId]
  );

  function setNotice(nextMessage) {
    setMessage(nextMessage);
    setError('');
  }

  function setFailure(nextError) {
    setError(nextError);
    setMessage('');
  }

  async function run(action, onSuccessMessage, callback) {
    setBusy(true);
    setError('');
    try {
      await callback();
      if (onSuccessMessage) setNotice(`${action}: ${onSuccessMessage}`);
    } catch (err) {
      setFailure(`${action} failed: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function loadMe() {
    const data = await gqlRequest(Q_ME);
    setMe(data.me);
    setMeForm({
      name: data.me.name || '',
      email: data.me.email || '',
      timezone: data.me.timezone || 'UTC',
      avatarUrl: data.me.avatarUrl || ''
    });
  }

  async function loadProjects() {
    const data = await gqlRequest(Q_PROJECTS, { first: 100 });
    const nextProjects = data.projects.edges.map((edge) => edge.node);
    setProjects(nextProjects);

    if (nextProjects.length === 0) {
      setSelectedProjectId('');
      setTasks([]);
      return;
    }

    if (!nextProjects.some((project) => project.id === selectedProjectId)) {
      const fallbackId = nextProjects[0].id;
      setSelectedProjectId(fallbackId);
      const fallback = nextProjects[0];
      setProjectUpdateForm({
        title: fallback.title || '',
        description: fallback.description || '',
        color: fallback.color || ''
      });
    }
  }

  async function loadLabels() {
    const data = await gqlRequest(Q_LABELS, { first: 200 });
    const nextLabels = data.labels.edges.map((edge) => edge.node);
    setLabels(nextLabels);
    if (selectedLabelId && !nextLabels.some((label) => label.id === selectedLabelId)) {
      setSelectedLabelId('');
      setLabelUpdateName('');
    }
  }

  async function loadTasks(projectId = selectedProjectId) {
    if (!projectId) {
      setTasks([]);
      return;
    }

    const data = await gqlRequest(Q_TASKS, {
      projectId,
      statuses: taskFilters.statuses,
      priorities: taskFilters.priorities,
      first: 200
    });
    const nextTasks = data.tasks.edges.map((edge) => edge.node);
    setTasks(nextTasks);

    if (selectedTaskId && !nextTasks.some((task) => task.id === selectedTaskId || task.subtasks?.some((sub) => sub.id === selectedTaskId))) {
      setSelectedTaskId('');
      resetTaskUpdateForm();
    }
  }

  async function refreshAll() {
    await loadMe();
    await loadProjects();
    await loadLabels();
  }

  useEffect(() => {
    run('initial load', 'loaded', refreshAll);
  }, []);

  useEffect(() => {
    if (!selectedProjectId) {
      setTasks([]);
      return;
    }
    run('tasks', 'loaded', () => loadTasks(selectedProjectId));
  }, [selectedProjectId]);

  function resetTaskUpdateForm() {
    setTaskUpdateForm({
      title: '',
      description: '',
      status: '',
      priority: '',
      startAt: '',
      dueAt: '',
      labelIds: []
    });
  }

  function syncProjectUpdateForm(projectId) {
    const project = projects.find((item) => item.id === projectId);
    if (!project) {
      setProjectUpdateForm({ title: '', description: '', color: '' });
      return;
    }
    setProjectUpdateForm({
      title: project.title || '',
      description: project.description || '',
      color: project.color || ''
    });
  }

  function syncTaskUpdateForm(taskId) {
    const task = allTasks.find((item) => item.id === taskId);
    if (!task) {
      resetTaskUpdateForm();
      return;
    }

    setTaskUpdateForm({
      title: task.title || '',
      description: task.description || '',
      status: task.status || '',
      priority: task.priority || '',
      startAt: toDatetimeLocal(task.startAt),
      dueAt: toDatetimeLocal(task.dueAt),
      labelIds: (task.labels || []).map((label) => label.id)
    });
  }

  function toggleMultiSelect(list, value) {
    return list.includes(value) ? list.filter((item) => item !== value) : [...list, value];
  }

  async function handleUpsertMe(event) {
    event.preventDefault();
    await run('upsertMe', 'saved', async () => {
      await gqlRequest(M_UPSERT_ME, {
        input: {
          name: meForm.name,
          email: meForm.email,
          timezone: meForm.timezone,
          avatarUrl: meForm.avatarUrl || null
        }
      });
      await loadMe();
    });
  }

  async function handleCreateProject(event) {
    event.preventDefault();
    await run('createProject', 'project created', async () => {
      await gqlRequest(M_CREATE_PROJECT, {
        input: {
          title: projectCreateForm.title,
          description: projectCreateForm.description || null,
          color: projectCreateForm.color || null
        }
      });
      setProjectCreateForm({ title: '', description: '', color: '' });
      await loadProjects();
      await loadTasks();
    });
  }

  async function handleUpdateProject(event) {
    event.preventDefault();
    if (!selectedProjectId) return;

    await run('updateProject', 'project updated', async () => {
      await gqlRequest(M_UPDATE_PROJECT, {
        input: {
          id: selectedProjectId,
          title: projectUpdateForm.title,
          description: projectUpdateForm.description || null,
          color: projectUpdateForm.color || null
        }
      });
      await loadProjects();
      await loadTasks();
    });
  }

  async function handleDeleteProject() {
    if (!selectedProjectId) return;
    await run('deleteProject', 'project deleted', async () => {
      await gqlRequest(M_DELETE_PROJECT, { id: selectedProjectId });
      setSelectedProjectId('');
      setTasks([]);
      await loadProjects();
    });
  }

  async function handleCreateLabel(event) {
    event.preventDefault();
    await run('createLabel', 'label created', async () => {
      await gqlRequest(M_CREATE_LABEL, { input: { name: labelCreateName } });
      setLabelCreateName('');
      await loadLabels();
    });
  }

  async function handleUpdateLabel(event) {
    event.preventDefault();
    if (!selectedLabelId) return;

    await run('updateLabel', 'label updated', async () => {
      await gqlRequest(M_UPDATE_LABEL, {
        input: {
          id: selectedLabelId,
          name: labelUpdateName
        }
      });
      await loadLabels();
      await loadTasks();
    });
  }

  async function handleDeleteLabel() {
    if (!selectedLabelId) return;

    await run('deleteLabel', 'label deleted', async () => {
      await gqlRequest(M_DELETE_LABEL, { id: selectedLabelId });
      setSelectedLabelId('');
      setLabelUpdateName('');
      await loadLabels();
      await loadTasks();
    });
  }

  async function handleCreateTask(event) {
    event.preventDefault();
    if (!selectedProjectId) return;

    await run('createTask', 'task created', async () => {
      await gqlRequest(M_CREATE_TASK, {
        input: {
          projectId: selectedProjectId,
          title: taskCreateForm.title,
          description: taskCreateForm.description || null,
          status: taskCreateForm.status,
          priority: taskCreateForm.priority,
          startAt: toISO(taskCreateForm.startAt),
          dueAt: toISO(taskCreateForm.dueAt),
          labelIds: taskCreateForm.labelIds
        }
      });
      setTaskCreateForm({
        title: '',
        description: '',
        status: 'TODO',
        priority: 'P3',
        startAt: '',
        dueAt: '',
        labelIds: []
      });
      await loadTasks();
    });
  }

  async function handleCreateSubtask(event) {
    event.preventDefault();
    if (!selectedProjectId || !subtaskCreateForm.parentTaskId) return;

    await run('createSubtask', 'subtask created', async () => {
      await gqlRequest(M_CREATE_TASK, {
        input: {
          projectId: selectedProjectId,
          parentTaskId: subtaskCreateForm.parentTaskId,
          title: subtaskCreateForm.title,
          description: subtaskCreateForm.description || null,
          status: subtaskCreateForm.status,
          priority: subtaskCreateForm.priority,
          startAt: toISO(subtaskCreateForm.startAt),
          dueAt: toISO(subtaskCreateForm.dueAt),
          labelIds: subtaskCreateForm.labelIds
        }
      });
      setSubtaskCreateForm({
        parentTaskId: '',
        title: '',
        description: '',
        status: 'TODO',
        priority: 'P3',
        startAt: '',
        dueAt: '',
        labelIds: []
      });
      await loadTasks();
    });
  }

  async function handleUpdateTask(event) {
    event.preventDefault();
    if (!selectedTaskId) return;

    const input = { id: selectedTaskId };

    if (taskUpdateForm.title.trim() !== '') input.title = taskUpdateForm.title;
    if (taskUpdateForm.description.trim() !== '') input.description = taskUpdateForm.description;
    if (taskUpdateForm.status) input.status = taskUpdateForm.status;
    if (taskUpdateForm.priority) input.priority = taskUpdateForm.priority;
    if (taskUpdateForm.startAt) input.startAt = toISO(taskUpdateForm.startAt);
    if (taskUpdateForm.dueAt) input.dueAt = toISO(taskUpdateForm.dueAt);
    if (taskUpdateForm.labelIds) input.labelIds = taskUpdateForm.labelIds;

    await run('updateTask', 'task updated', async () => {
      await gqlRequest(M_UPDATE_TASK, { input });
      await loadTasks();
      syncTaskUpdateForm(selectedTaskId);
    });
  }

  async function handleDeleteTask() {
    if (!selectedTaskId) return;

    await run('deleteTask', 'task deleted', async () => {
      await gqlRequest(M_DELETE_TASK, { id: selectedTaskId });
      setSelectedTaskId('');
      resetTaskUpdateForm();
      await loadTasks();
    });
  }

  function renderLabelChips(taskLabels = []) {
    if (!taskLabels.length) return <span className="tag">no labels</span>;
    return taskLabels.map((label) => (
      <span className="tag" key={label.id}>
        {label.name}
      </span>
    ));
  }

  function toggleTaskFilter(group, value) {
    setTaskFilters((current) => ({
      ...current,
      [group]: toggleMultiSelect(current[group], value)
    }));
  }

  async function applyFilters() {
    await run('tasks filter', 'tasks refreshed', () => loadTasks());
  }

  return (
    <main className="page">
      <header className="header">
        <h1>ZenList Backend API Tester</h1>
        <p>GraphQL URL: {GRAPHQL_URL}</p>
        <div className="actions-row">
          <button disabled={busy} onClick={() => run('refresh', 'all resources refreshed', refreshAll)}>
            Refresh everything
          </button>
          <button disabled={busy || !selectedProjectId} onClick={() => run('reload tasks', 'tasks refreshed', () => loadTasks())}>
            Refresh tasks
          </button>
        </div>
        {message ? <p className="success">{message}</p> : null}
        {error ? <p className="error">{error}</p> : null}
      </header>

      <section className="panel">
        <h2>User Profile (`upsertMe`, `me`)</h2>
        <form onSubmit={handleUpsertMe} className="form-grid">
          <label>
            Name
            <input value={meForm.name} onChange={(event) => setMeForm((prev) => ({ ...prev, name: event.target.value }))} />
          </label>
          <label>
            Email
            <input value={meForm.email} onChange={(event) => setMeForm((prev) => ({ ...prev, email: event.target.value }))} />
          </label>
          <label>
            Timezone
            <input value={meForm.timezone} onChange={(event) => setMeForm((prev) => ({ ...prev, timezone: event.target.value }))} />
          </label>
          <label>
            Avatar URL
            <input value={meForm.avatarUrl} onChange={(event) => setMeForm((prev) => ({ ...prev, avatarUrl: event.target.value }))} />
          </label>
          <button disabled={busy} type="submit">
            Save profile
          </button>
        </form>
        <pre>{JSON.stringify(me, null, 2)}</pre>
      </section>

      <section className="panel two-col">
        <div>
          <h2>Projects</h2>
          <form onSubmit={handleCreateProject} className="form-grid">
            <label>
              Title
              <input
                value={projectCreateForm.title}
                onChange={(event) => setProjectCreateForm((prev) => ({ ...prev, title: event.target.value }))}
                required
              />
            </label>
            <label>
              Description
              <input
                value={projectCreateForm.description}
                onChange={(event) => setProjectCreateForm((prev) => ({ ...prev, description: event.target.value }))}
              />
            </label>
            <label>
              Color (#RRGGBB)
              <input value={projectCreateForm.color} onChange={(event) => setProjectCreateForm((prev) => ({ ...prev, color: event.target.value }))} />
            </label>
            <button disabled={busy} type="submit">
              Create project
            </button>
          </form>

          <label>
            Select Project
            <select
              value={selectedProjectId}
              onChange={(event) => {
                const nextId = event.target.value;
                setSelectedProjectId(nextId);
                syncProjectUpdateForm(nextId);
              }}
            >
              <option value="">-- choose --</option>
              {projects.map((project) => (
                <option key={project.id} value={project.id}>
                  {project.title}
                </option>
              ))}
            </select>
          </label>

          <form onSubmit={handleUpdateProject} className="form-grid">
            <label>
              Title
              <input
                value={projectUpdateForm.title}
                onChange={(event) => setProjectUpdateForm((prev) => ({ ...prev, title: event.target.value }))}
                required
              />
            </label>
            <label>
              Description
              <input
                value={projectUpdateForm.description}
                onChange={(event) => setProjectUpdateForm((prev) => ({ ...prev, description: event.target.value }))}
              />
            </label>
            <label>
              Color
              <input value={projectUpdateForm.color} onChange={(event) => setProjectUpdateForm((prev) => ({ ...prev, color: event.target.value }))} />
            </label>
            <div className="actions-row">
              <button disabled={busy || !selectedProjectId} type="submit">
                Update project
              </button>
              <button disabled={busy || !selectedProjectId} type="button" onClick={handleDeleteProject}>
                Delete project
              </button>
            </div>
          </form>
        </div>

        <div>
          <h3>Projects List</h3>
          <pre>{JSON.stringify(projects, null, 2)}</pre>
        </div>

        <div className="panel-footer">
          <a href="https://github.com/faizp/RealtimeKit" target="_blank" rel="noopener noreferrer" className="github-link" title="View on GitHub">
            <svg width="20" height="20" viewBox="0 0 16 16" fill="currentColor">
              <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
            </svg>
          </a>
        </div>
      </section>

      <section className="panel two-col">
        <div>
          <h2>Labels</h2>
          <form onSubmit={handleCreateLabel} className="form-grid">
            <label>
              Label Name
              <input value={labelCreateName} onChange={(event) => setLabelCreateName(event.target.value)} required />
            </label>
            <button disabled={busy} type="submit">
              Create label
            </button>
          </form>

          <label>
            Select Label
            <select
              value={selectedLabelId}
              onChange={(event) => {
                const nextId = event.target.value;
                setSelectedLabelId(nextId);
                const selected = labels.find((label) => label.id === nextId);
                setLabelUpdateName(selected?.name || '');
              }}
            >
              <option value="">-- choose --</option>
              {labels.map((label) => (
                <option key={label.id} value={label.id}>
                  {label.name}
                </option>
              ))}
            </select>
          </label>

          <form onSubmit={handleUpdateLabel} className="form-grid">
            <label>
              Update Name
              <input value={labelUpdateName} onChange={(event) => setLabelUpdateName(event.target.value)} required />
            </label>
            <div className="actions-row">
              <button disabled={busy || !selectedLabel} type="submit">
                Update label
              </button>
              <button disabled={busy || !selectedLabel} type="button" onClick={handleDeleteLabel}>
                Delete label
              </button>
            </div>
          </form>
        </div>

        <div>
          <h3>Labels List</h3>
          <pre>{JSON.stringify(labels, null, 2)}</pre>
        </div>
      </section>

      <section className="panel">
        <h2>Tasks + Subtasks (Project: {selectedProject?.title || 'none'})</h2>

        <div className="filters">
          <strong>Filter Status:</strong>
          {STATUS_OPTIONS.map((status) => (
            <label key={`status-${status}`}>
              <input
                type="checkbox"
                checked={taskFilters.statuses.includes(status)}
                onChange={() => toggleTaskFilter('statuses', status)}
              />
              {status}
            </label>
          ))}
          <strong>Filter Priority:</strong>
          {PRIORITY_OPTIONS.map((priority) => (
            <label key={`priority-${priority}`}>
              <input
                type="checkbox"
                checked={taskFilters.priorities.includes(priority)}
                onChange={() => toggleTaskFilter('priorities', priority)}
              />
              {priority}
            </label>
          ))}
          <button disabled={busy || !selectedProjectId} type="button" onClick={applyFilters}>
            Apply filters
          </button>
        </div>

        <div className="three-col">
          <form onSubmit={handleCreateTask} className="form-grid">
            <h3>Create Root Task</h3>
            <label>
              Title
              <input value={taskCreateForm.title} onChange={(event) => setTaskCreateForm((prev) => ({ ...prev, title: event.target.value }))} required />
            </label>
            <label>
              Description
              <input
                value={taskCreateForm.description}
                onChange={(event) => setTaskCreateForm((prev) => ({ ...prev, description: event.target.value }))}
              />
            </label>
            <label>
              Status
              <select value={taskCreateForm.status} onChange={(event) => setTaskCreateForm((prev) => ({ ...prev, status: event.target.value }))}>
                {STATUS_OPTIONS.map((status) => (
                  <option key={status} value={status}>
                    {status}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Priority
              <select value={taskCreateForm.priority} onChange={(event) => setTaskCreateForm((prev) => ({ ...prev, priority: event.target.value }))}>
                {PRIORITY_OPTIONS.map((priority) => (
                  <option key={priority} value={priority}>
                    {priority}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Start At
              <input
                type="datetime-local"
                value={taskCreateForm.startAt}
                onChange={(event) => setTaskCreateForm((prev) => ({ ...prev, startAt: event.target.value }))}
              />
            </label>
            <label>
              Due At
              <input
                type="datetime-local"
                value={taskCreateForm.dueAt}
                onChange={(event) => setTaskCreateForm((prev) => ({ ...prev, dueAt: event.target.value }))}
              />
            </label>
            <fieldset>
              <legend>Labels</legend>
              {labels.map((label) => (
                <label key={`create-task-label-${label.id}`}>
                  <input
                    type="checkbox"
                    checked={taskCreateForm.labelIds.includes(label.id)}
                    onChange={() =>
                      setTaskCreateForm((prev) => ({
                        ...prev,
                        labelIds: toggleMultiSelect(prev.labelIds, label.id)
                      }))
                    }
                  />
                  {label.name}
                </label>
              ))}
            </fieldset>
            <button disabled={busy || !selectedProjectId} type="submit">
              Create root task
            </button>
          </form>

          <form onSubmit={handleCreateSubtask} className="form-grid">
            <h3>Create Subtask</h3>
            <label>
              Parent Root Task
              <select
                value={subtaskCreateForm.parentTaskId}
                onChange={(event) => setSubtaskCreateForm((prev) => ({ ...prev, parentTaskId: event.target.value }))}
                required
              >
                <option value="">-- choose parent --</option>
                {rootTasks.map((task) => (
                  <option key={task.id} value={task.id}>
                    {task.title}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Title
              <input
                value={subtaskCreateForm.title}
                onChange={(event) => setSubtaskCreateForm((prev) => ({ ...prev, title: event.target.value }))}
                required
              />
            </label>
            <label>
              Description
              <input
                value={subtaskCreateForm.description}
                onChange={(event) => setSubtaskCreateForm((prev) => ({ ...prev, description: event.target.value }))}
              />
            </label>
            <label>
              Status
              <select value={subtaskCreateForm.status} onChange={(event) => setSubtaskCreateForm((prev) => ({ ...prev, status: event.target.value }))}>
                {STATUS_OPTIONS.map((status) => (
                  <option key={status} value={status}>
                    {status}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Priority
              <select value={subtaskCreateForm.priority} onChange={(event) => setSubtaskCreateForm((prev) => ({ ...prev, priority: event.target.value }))}>
                {PRIORITY_OPTIONS.map((priority) => (
                  <option key={priority} value={priority}>
                    {priority}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Start At
              <input
                type="datetime-local"
                value={subtaskCreateForm.startAt}
                onChange={(event) => setSubtaskCreateForm((prev) => ({ ...prev, startAt: event.target.value }))}
              />
            </label>
            <label>
              Due At
              <input
                type="datetime-local"
                value={subtaskCreateForm.dueAt}
                onChange={(event) => setSubtaskCreateForm((prev) => ({ ...prev, dueAt: event.target.value }))}
              />
            </label>
            <fieldset>
              <legend>Labels</legend>
              {labels.map((label) => (
                <label key={`create-subtask-label-${label.id}`}>
                  <input
                    type="checkbox"
                    checked={subtaskCreateForm.labelIds.includes(label.id)}
                    onChange={() =>
                      setSubtaskCreateForm((prev) => ({
                        ...prev,
                        labelIds: toggleMultiSelect(prev.labelIds, label.id)
                      }))
                    }
                  />
                  {label.name}
                </label>
              ))}
            </fieldset>
            <button disabled={busy || !selectedProjectId} type="submit">
              Create subtask
            </button>
          </form>

          <form onSubmit={handleUpdateTask} className="form-grid">
            <h3>Update / Delete Task</h3>
            <label>
              Select Task
              <select
                value={selectedTaskId}
                onChange={(event) => {
                  const nextId = event.target.value;
                  setSelectedTaskId(nextId);
                  syncTaskUpdateForm(nextId);
                }}
              >
                <option value="">-- choose task --</option>
                {allTasks.map((task) => (
                  <option key={task.id} value={task.id}>
                    {task.parentTaskId ? 'Subtask: ' : 'Task: '}
                    {task.title}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Title
              <input value={taskUpdateForm.title} onChange={(event) => setTaskUpdateForm((prev) => ({ ...prev, title: event.target.value }))} />
            </label>
            <label>
              Description
              <input
                value={taskUpdateForm.description}
                onChange={(event) => setTaskUpdateForm((prev) => ({ ...prev, description: event.target.value }))}
              />
            </label>
            <label>
              Status
              <select value={taskUpdateForm.status} onChange={(event) => setTaskUpdateForm((prev) => ({ ...prev, status: event.target.value }))}>
                <option value="">(no change)</option>
                {STATUS_OPTIONS.map((status) => (
                  <option key={status} value={status}>
                    {status}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Priority
              <select value={taskUpdateForm.priority} onChange={(event) => setTaskUpdateForm((prev) => ({ ...prev, priority: event.target.value }))}>
                <option value="">(no change)</option>
                {PRIORITY_OPTIONS.map((priority) => (
                  <option key={priority} value={priority}>
                    {priority}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Start At
              <input
                type="datetime-local"
                value={taskUpdateForm.startAt}
                onChange={(event) => setTaskUpdateForm((prev) => ({ ...prev, startAt: event.target.value }))}
              />
            </label>
            <label>
              Due At
              <input
                type="datetime-local"
                value={taskUpdateForm.dueAt}
                onChange={(event) => setTaskUpdateForm((prev) => ({ ...prev, dueAt: event.target.value }))}
              />
            </label>
            <fieldset>
              <legend>Labels</legend>
              {labels.map((label) => (
                <label key={`update-task-label-${label.id}`}>
                  <input
                    type="checkbox"
                    checked={taskUpdateForm.labelIds.includes(label.id)}
                    onChange={() =>
                      setTaskUpdateForm((prev) => ({
                        ...prev,
                        labelIds: toggleMultiSelect(prev.labelIds, label.id)
                      }))
                    }
                  />
                  {label.name}
                </label>
              ))}
            </fieldset>
            <div className="actions-row">
              <button disabled={busy || !selectedTaskId} type="submit">
                Update task
              </button>
              <button disabled={busy || !selectedTaskId} type="button" onClick={handleDeleteTask}>
                Delete task
              </button>
            </div>
          </form>
        </div>
      </section>

      <section className="panel">
        <h3>Task Snapshot</h3>
        {rootTasks.length === 0 ? (
          <p>No tasks for this project/filter.</p>
        ) : (
          <div className="task-list">
            {rootTasks.map((task) => (
              <article key={task.id} className="task-card">
                <h4>
                  {task.title} ({task.status} / {task.priority})
                </h4>
                <p>{task.description || 'No description'}</p>
                <p>
                  Start: {formatDate(task.startAt)} | Due: {formatDate(task.dueAt)} | Completed: {formatDate(task.completedAt)}
                </p>
                <div className="chip-row">{renderLabelChips(task.labels)}</div>
                <div className="subtasks">
                  <strong>Subtasks ({task.subtasks?.length || 0})</strong>
                  {(task.subtasks || []).map((subtask) => (
                    <div key={subtask.id} className="subtask-item">
                      <span>
                        {subtask.title} ({subtask.status}/{subtask.priority})
                      </span>
                      <div className="chip-row">{renderLabelChips(subtask.labels)}</div>
                    </div>
                  ))}
                </div>
              </article>
            ))}
          </div>
        )}
      </section>
    </main>
  );
}
